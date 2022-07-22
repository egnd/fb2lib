package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/profile"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/factories"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/go-pipeline/pools"
	"github.com/egnd/go-pipeline/tasks"
)

var (
	appVersion = "debug"

	showVersion = flag.Bool("version", false, "Show app version.")
	hideBar     = flag.Bool("hidebar", false, "Hide progress bar.")
	batchSize   = flag.Int("batch", 1000, "Books batch size.")
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "FBL", "Prefix for env variables.")
	profiler    = flag.String("pprof", "", "Enable profiler (mem,allocs,heap,cpu,trace,goroutine,mutex,block,thread).")

	resetStats = func() map[repos.BucketType]entities.ItemFreqMap {
		return map[repos.BucketType]entities.ItemFreqMap{
			repos.BucketGenres:  make(entities.ItemFreqMap, *batchSize),
			repos.BucketAuthors: make(entities.ItemFreqMap, *batchSize),
			repos.BucketSeries:  make(entities.ItemFreqMap, *batchSize),
			repos.BucketLangs:   make(entities.ItemFreqMap, *batchSize),
			repos.BucketLibs:    make(entities.ItemFreqMap, *batchSize),
		}
	}
)

func main() {
	var (
		step     int
		cntTotal uint64
		barTotal *mpb.Bar
		startTS  = time.Now()
	)

	flag.Parse()

	if *showVersion {
		fmt.Println(appVersion)
		return
	}

	cfg := factories.NewViperCfg(*cfgPath, *cfgPrefix)
	logger := factories.NewZerolog(cfg, os.Stderr)

	defer func() {
		logger.Info().Uint64("cnt", cntTotal).Dur("dur", time.Since(startTS)).Msg("summary build finished")
	}()

	if *profiler != "" {
		defer RunProfiler(*profiler, cfg).Stop()
	}

	os.RemoveAll(path.Join(cfg.GetString("adapters.badger.dir"), "authors"))
	os.RemoveAll(path.Join(cfg.GetString("adapters.badger.dir"), "series"))
	os.RemoveAll(path.Join(cfg.GetString("adapters.badger.dir"), "genres"))
	os.RemoveAll(path.Join(cfg.GetString("adapters.badger.dir"), "libs"))
	os.RemoveAll(path.Join(cfg.GetString("adapters.badger.dir"), "langs"))

	repoBooks := repos.NewBooksLevelBleve(0,
		map[repos.BucketType]*leveldb.DB{
			repos.BucketBooks:   factories.NewLevelDB(cfg.GetString("adapters.badger.dir"), "books"),
			repos.BucketAuthors: factories.NewLevelDB(cfg.GetString("adapters.badger.dir"), "authors"),
			repos.BucketSeries:  factories.NewLevelDB(cfg.GetString("adapters.badger.dir"), "series"),
			repos.BucketGenres:  factories.NewLevelDB(cfg.GetString("adapters.badger.dir"), "genres"),
			repos.BucketLibs:    factories.NewLevelDB(cfg.GetString("adapters.badger.dir"), "libs"),
			repos.BucketLangs:   factories.NewLevelDB(cfg.GetString("adapters.badger.dir"), "langs"),
		},
		factories.NewBleveIndex(cfg.GetString("adapters.bleve.dir"), "books", entities.NewBookIndexMapping()),
		jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
		logger,
	)
	defer repoBooks.Close()

	if !*hideBar {
		barTotal = GetProgressBar(mpb.New(mpb.WithOutput(os.Stdout)), cfg, &logger, repoBooks)
	}

	stats := resetStats()

	pipe := pools.NewSemaphore(len(stats), nil)
	defer pipe.Close()

	if err := repoBooks.IterateOver(func(book *entities.Book) error {
		defer func() {
			if barTotal != nil {
				barTotal.Increment()
			}
		}()

		logger.Debug().Str("book", book.ID).Msg("calculate")

		for _, k := range book.Genres() {
			stats[repos.BucketGenres].Put(k, 1)
		}
		for _, k := range book.Authors() {
			stats[repos.BucketAuthors].Put(k, 1)
		}
		for _, k := range book.Series() {
			stats[repos.BucketSeries].Put(k, 1)
		}
		stats[repos.BucketLangs].Put(book.Info.Lang, 1)
		stats[repos.BucketLibs].Put(book.Lib, 1)

		step++
		cntTotal++
		if step < *batchSize {
			return nil
		}

		for bucket, freq := range stats {
			bucket, freq := bucket, freq
			pipe.Push(tasks.NewFunc(string(bucket), func() error {
				if err := repoBooks.AppendFreqs(bucket, freq); err != nil {
					logger.Warn().Err(err).Str("bucket", string(bucket)).Int("len", len(freq)).Msg("batch")
				} else {
					logger.Debug().Str("bucket", string(bucket)).Int("len", len(freq)).Msg("batch")
				}

				return nil
			}))
		}
		pipe.Wait()
		stats = resetStats()
		step = 0
		return nil
	}); err != nil {
		logger.Error().Err(err).Msg("iterating over books")
	}

	for bucket, freq := range stats {
		bucket, freq := bucket, freq
		pipe.Push(tasks.NewFunc(string(bucket), func() error {
			if err := repoBooks.AppendFreqs(bucket, freq); err != nil {
				logger.Warn().Err(err).Str("bucket", string(bucket)).Int("len", len(freq)).Msg("batch")
			} else {
				logger.Debug().Str("bucket", string(bucket)).Int("len", len(freq)).Msg("batch")
			}

			return nil
		}))
	}
	pipe.Wait()
	time.Sleep(100 * time.Millisecond)
}

func RunProfiler(profType string, cfg *viper.Viper) interface{ Stop() } {
	_ = os.MkdirAll(cfg.GetString("pprof.dir"), 0755)

	pprofopts := []func(*profile.Profile){
		profile.ProfilePath(cfg.GetString("pprof.dir")),
		profile.NoShutdownHook,
	}

	switch profType {
	case "mem":
		pprofopts = append(pprofopts, profile.MemProfile)
	case "allocs":
		pprofopts = append(pprofopts, profile.MemProfileAllocs)
	case "heap":
		pprofopts = append(pprofopts, profile.MemProfileHeap)
	case "cpu":
		pprofopts = append(pprofopts, profile.CPUProfile)
	case "trace":
		pprofopts = append(pprofopts, profile.TraceProfile)
	case "goroutine":
		pprofopts = append(pprofopts, profile.GoroutineProfile)
	case "mutex":
		pprofopts = append(pprofopts, profile.MutexProfile)
	case "block":
		pprofopts = append(pprofopts, profile.BlockProfile)
	case "thread":
		pprofopts = append(pprofopts, profile.ThreadcreationProfile)
	default:
		log.Fatal().Str("type", profType).Msg("invalid profiling type")
	}

	return profile.Start(pprofopts...)
}

func GetProgressBar(bars *mpb.Progress, cfg *viper.Viper, logger *zerolog.Logger, repo *repos.BooksLevelBleve) *mpb.Bar {
	if err := os.MkdirAll(cfg.GetString("logs.dir"), 0644); err != nil {
		panic(err)
	}

	logOutput, err := os.OpenFile(fmt.Sprintf("%s/indexing.%d.log",
		cfg.GetString("logs.dir"), time.Now().Unix(),
	), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	*logger = logger.Output(zerolog.ConsoleWriter{Out: logOutput, NoColor: true})

	return bars.AddBar(int64(repo.GetTotal()),
		mpb.PrependDecorators(
			decor.Name("summary "),
			decor.Elapsed(decor.ET_STYLE_GO),
		),
		mpb.AppendDecorators(
			decor.AverageETA(decor.ET_STYLE_GO),
		),
	)
}
