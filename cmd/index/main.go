package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/pools"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/profile"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/factories"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/fb2lib/internal/tasks"
)

var (
	appVersion = "debug"

	showVersion = flag.Bool("version", false, "Show app version.")
	hideBar     = flag.Bool("hidebar", false, "Hide progress bar.")
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "FBL", "Prefix for env variables.")
	profiler    = flag.String("pprof", "", "Enable profiler (mem,allocs,heap,cpu,trace,goroutine,mutex,block,thread).")
)

func main() {
	var (
		wg         sync.WaitGroup
		bars       *mpb.Progress
		barTotal   *mpb.Bar
		cntTotal   entities.CntAtomic32
		cntIndexed entities.CntAtomic32
		startTS    = time.Now()
	)

	flag.Parse()

	if *showVersion {
		fmt.Println(appVersion)
		return
	}

	cfg := InitConfig(*cfgPath, *cfgPrefix)
	logger := factories.NewZerolog(cfg, os.Stderr)

	defer func() {
		logger.Info().
			Uint32("succeed", cntIndexed.Total()).
			Uint32("failed", cntTotal.Total()-cntIndexed.Total()).
			Dur("dur", time.Since(startTS)).Msg("indexing finished")
	}()

	if *profiler != "" {
		defer RunProfiler(*profiler, cfg).Stop()
	}

	libs, err := entities.NewLibraries("libraries", cfg)
	if err != nil {
		panic(err)
	}

	if !*hideBar {
		bars = mpb.New(mpb.WithOutput(os.Stdout))
		barTotal = GetProgressBar(bars, libs, cfg, &logger)
	}

	repoBooks := repos.NewBooksBadgerBleve(cfg.GetInt("indexer.batch_size"),
		map[repos.BucketType]*badger.DB{
			repos.BucketBooks:   factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "books"),
			repos.BucketAuthors: factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "authors"),
			repos.BucketSeries:  factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "series"),
			repos.BucketGenres:  factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "genres"),
			repos.BucketLibs:    factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "libs"),
			repos.BucketLangs:   factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "langs"),
		},
		factories.NewBleveIndex(cfg.GetString("adapters.bleve.dir"), "books", entities.NewBookIndexMapping()),
		jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
		logger,
	)
	defer repoBooks.Close()

	repoMarks := repos.NewLibMarks(factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "marks"))
	defer repoMarks.Close()

	rules, err := entities.NewIndexRules("indexer.rules", cfg)
	if err != nil {
		panic(err)
	}

	pipe := pools.NewSemaphore(cfg.GetInt("indexer.threads_cnt"), &wg, func(next pipeline.TaskExecutor) pipeline.TaskExecutor {
		return func(task pipeline.Task) error {
			logger.Debug().Str("task", task.ID()).Msg("iterate")

			if err := next(task); err != nil {
				var indexed *tasks.ErrAlreadyIndexed
				if errors.As(err, &indexed) {
					logger.Info().Str("task", task.ID()).Msg(err.Error())
				} else {
					logger.Error().Str("task", task.ID()).Err(err).Msg("iterate")
				}
			}

			return nil
		}
	})
	defer pipe.Close()

	readingPool := pools.NewBusPool(cfg.GetInt("indexer.read_threads"), cfg.GetInt("indexer.read_buff"), &wg,
		func(next pipeline.TaskExecutor) pipeline.TaskExecutor {
			return func(task pipeline.Task) error {
				logger.Debug().Str("task", task.ID()).Msg("read")

				if err := next(task); err != nil {
					logger.Error().Str("task", task.ID()).Err(err).Msg("read")
				}

				return nil
			}
		},
	)
	defer readingPool.Close()

	parsingPool := pools.NewBusPool(cfg.GetInt("indexer.parse_threads"), cfg.GetInt("indexer.parse_buff"), &wg,
		func(next pipeline.TaskExecutor) pipeline.TaskExecutor {
			return func(task pipeline.Task) error {
				logger.Debug().Str("task", task.ID()).Msg("parse")

				if err := next(task); err != nil {
					var skipRule *tasks.ErrSkipRule
					if errors.As(err, &skipRule) {
						logger.Warn().Str("task", task.ID()).Msg(err.Error())
					} else {
						logger.Error().Str("task", task.ID()).Err(err).Msg("parse")
					}
				} else {
					cntIndexed.Inc(1)
				}

				return nil
			}
		},
	)
	defer parsingPool.Close()

	libItems, err := libs.GetItems()
	if err != nil {
		panic(err)
	}

	var num int
	total := len(libItems)
	for _, v := range libItems {
		num++
		lib := libs[v.Lib]
		itemPath := v.Item
		readerTaskFactory := func(reader io.ReadCloser, book entities.Book) error {
			cntTotal.Inc(1)
			return readingPool.Push(tasks.NewReadTask(book.Src, book.Lib, reader, func(data io.Reader) error {
				return parsingPool.Push(tasks.NewParseFB2Task(data, book, rules, lib.Encoder, repoBooks, barTotal))
			}))
		}
		pipe.Push(tasks.NewDefineItemTask(itemPath, lib, repoMarks, readerTaskFactory, func(finfo fs.FileInfo) error {
			return tasks.NewReadZipTask(num, total, itemPath, finfo, lib, bars, readerTaskFactory).Do()
		}))
	}

	wg.Wait()
}

func InitConfig(path, prefix string) *viper.Viper {
	cfg := factories.NewViperCfg(*cfgPath, *cfgPrefix)

	if cfg.GetInt("indexer.threads_cnt") == 0 {
		cfg.Set("indexer.threads_cnt", 1)
	}

	if cfg.GetInt("indexer.read_threads") == 0 {
		cfg.Set("indexer.read_threads", cfg.GetInt("indexer.threads_cnt")*10)
	}

	if cfg.GetInt("indexer.parse_threads") == 0 {
		cnt := runtime.NumCPU() / 2
		if cnt == 0 {
			cnt = 1
		}

		cfg.Set("indexer.parse_threads", cnt)
	}

	return cfg
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

func GetProgressBar(bars *mpb.Progress, libs entities.Libraries, cfg *viper.Viper, logger *zerolog.Logger) *mpb.Bar {
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

	return bars.AddBar(libs.GetSize(),
		// mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),
		mpb.PrependDecorators(
			decor.Elapsed(decor.ET_STYLE_GO),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f/% .2f"), decor.Name(", "),
			decor.AverageSpeed(decor.UnitKB, "% .2f"), decor.Name(", "),
			decor.AverageETA(decor.ET_STYLE_GO),
		),
	)
}
