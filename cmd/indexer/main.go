package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"sync"
	"time"

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
	cfgPrefix   = flag.String("env-prefix", "BS", "Prefix for env variables.")
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

	if *profiler != "" {
		defer RunProfiler(*profiler, cfg).Stop()
	}

	logger := factories.NewZerolog(cfg, os.Stderr)
	libs := GetLibs(cfg, logger)

	if !*hideBar {
		bars = mpb.New(mpb.WithOutput(os.Stdout))
		barTotal = GetProgressBar(bars, libs, cfg, &logger)
	}

	repoBooks := repos.NewBooksBadgerBleve(cfg.GetInt("indexer.batch_size"),
		factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "books"),
		factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "authors"),
		factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "series"),
		factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "genres"),
		factories.NewBleveIndex(cfg.GetString("adapters.bleve.dir"), "books", entities.NewBookIndexMapping()),
		jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
		logger,
	)
	defer repoBooks.Close()

	repoMarks := repos.NewLibMarks(factories.NewBadgerDB(cfg.GetString("adapters.badger.dir"), "marks"))
	defer repoMarks.Close()

	pipe := pools.NewSemaphore(cfg.GetInt("indexer.threads_cnt"), func(next pipeline.TaskExecutor) pipeline.TaskExecutor {
		return func(task pipeline.Task) error {
			logger.Debug().Str("task", task.ID()).Msg("iterate")

			if err := next(task); err != nil {
				logger.Error().Str("task", task.ID()).Err(err).Msg("iterate")
			}

			wg.Done()

			return nil
		}
	})
	defer pipe.Close()

	readingPool := pools.NewBusPool(cfg.GetInt("indexer.read_threads"), cfg.GetInt("indexer.read_buff"),
		func(next pipeline.TaskExecutor) pipeline.TaskExecutor {
			return func(task pipeline.Task) error {
				logger.Debug().Str("task", task.ID()).Msg("read")

				if err := next(task); err != nil {
					logger.Error().Str("task", task.ID()).Err(err).Msg("read")
				}

				wg.Done()

				return nil
			}
		},
	)
	defer readingPool.Close()

	parsingPool := pools.NewBusPool(cfg.GetInt("indexer.parse_threads"), cfg.GetInt("indexer.parse_buff"),
		func(next pipeline.TaskExecutor) pipeline.TaskExecutor {
			return func(task pipeline.Task) error {
				logger.Debug().Str("task", task.ID()).Msg("parse")

				if err := next(task); err != nil {
					logger.Error().Str("task", task.ID()).Err(err).Msg("parse")
				} else {
					cntIndexed.Inc(1)
				}

				wg.Done()

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
		lib := libs[v.Lib]
		libItemPath := v.Item
		readerTaskFactory := func(reader io.ReadCloser, book entities.Book) error {
			wg.Add(1)
			cntTotal.Inc(1)
			return readingPool.Push(tasks.NewReadTask(book.Src, book.Lib, reader, func(data io.Reader) error {
				wg.Add(1)
				return parsingPool.Push(tasks.NewParseFB2Task(data, book, lib.Encoder, repoBooks, barTotal))
			}))
		}

		num++
		wg.Add(1)
		pipe.Push(tasks.NewDefineItemTask(libItemPath, lib, repoMarks, readerTaskFactory, func(finfo fs.FileInfo) error {
			return tasks.NewReadZipTask(num, total, libItemPath, finfo, lib, bars, readerTaskFactory).Do()
		}))
	}

	wg.Wait()

	logger.Info().
		Uint32("succeed", cntIndexed.Total()).
		Uint32("failed", cntTotal.Total()-cntIndexed.Total()).
		Dur("dur", time.Since(startTS)).Msg("indexing finished")
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

func GetLibs(cfg *viper.Viper, logger zerolog.Logger) entities.Libraries {
	res, err := entities.NewLibraries("libraries", cfg)
	if err != nil {
		panic(err)
	}

	return res
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
