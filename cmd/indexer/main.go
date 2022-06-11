package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/pool"
	"github.com/egnd/go-pipeline/semaphore"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/profile"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"go.etcd.io/bbolt"

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
	libsNames   = flag.String("libs", "", "Limit libraries to process.")
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

	storage := factories.NewBoltDB(cfg.GetString("boltdb.path"))
	defer storage.Close()

	logger := factories.NewZerolog(cfg, os.Stderr)
	libs := GetLibs(*libsNames, cfg, logger)

	if !*hideBar {
		bars = mpb.New(mpb.WithOutput(os.Stdout))
		barTotal = GetProgressBar(bars, libs, cfg, &logger)
	}

	indexList := map[string]bleve.Index{}
	reposInfo := GetInfoRepos(cfg.GetInt("indexer.batch_size"), libs, logger, cfg, storage, indexList)
	repoMarks := repos.NewLibMarks("indexed_items", storage)

	pipe := semaphore.NewSemaphore(cfg.GetInt("indexer.threads_cnt"), func(next pipeline.Tasker) pipeline.Tasker {
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

	readingPipeline := make(chan pipeline.Doer, cfg.GetInt("indexer.read_buff"))
	defer close(readingPipeline)

	readingPool := GetPool(cfg.GetInt("indexer.parse_threads"), readingPipeline, func(next pipeline.Tasker) pipeline.Tasker {
		return func(task pipeline.Task) error {
			logger.Debug().Str("task", task.ID()).Msg("read")

			if err := next(task); err != nil {
				logger.Error().Str("task", task.ID()).Err(err).Msg("read")
			}

			wg.Done()

			return nil
		}
	})
	defer readingPool.Close()

	parsingPipeline := make(chan pipeline.Doer, cfg.GetInt("indexer.parse_buff"))
	defer close(parsingPipeline)

	parsingPool := GetPool(cfg.GetInt("indexer.read_threads"), parsingPipeline, func(next pipeline.Tasker) pipeline.Tasker {
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
	})
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
		readerTaskFactory := func(reader io.ReadCloser, book entities.BookInfo) error {
			wg.Add(1)
			cntTotal.Inc(1)
			return readingPool.Push(tasks.NewReadTask(book.Src, book.LibName, reader, func(data io.Reader) error {
				wg.Add(1)
				return parsingPool.Push(tasks.NewParseFB2Task(data, book, lib.Encoder, reposInfo[lib.Name], barTotal))
			}))
		}

		num++
		wg.Add(1)
		pipe.Push(tasks.NewDefineItemTask(libItemPath, lib, repoMarks, readerTaskFactory, func(finfo fs.FileInfo) error {
			return tasks.NewReadZipTask(num, total, libItemPath, finfo, lib, bars, readerTaskFactory).Do()
		}))
	}

	wg.Wait()

	closedRepos := make(map[entities.IBooksInfoRepo]struct{}, len(reposInfo))
	for _, item := range reposInfo {
		if _, ok := closedRepos[item]; ok {
			continue
		}
		closedRepos[item] = struct{}{}
		item.Close()
	}

	for _, item := range indexList {
		if err := item.Close(); err != nil {
			panic(err)
		}
	}

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

func GetLibs(libsNames string, cfg *viper.Viper, logger zerolog.Logger) entities.Libraries {
	res, err := entities.NewLibraries("libraries", cfg)
	if err != nil {
		panic(err)
	}

	libs := strings.Split(libsNames, ",")
	if len(libs) < 2 && strings.TrimSpace(libs[0]) == "" {
		return res
	}

	for k, lib := range res {
		lib.Disabled = true
		res[k] = lib
	}

	for _, lib := range libs {
		if lib = strings.TrimSpace(lib); lib == "" {
			continue
		}

		libItem := res[lib]
		libItem.Disabled = false
		res[libItem.Name] = libItem
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

func GetInfoRepos(batchSize int, libs entities.Libraries, logger zerolog.Logger, cfg *viper.Viper, storage *bbolt.DB, indexList map[string]bleve.Index) map[string]entities.IBooksInfoRepo {
	res := make(map[string]entities.IBooksInfoRepo, len(libs))

	for _, lib := range libs {
		if lib.Disabled {
			continue
		}

		if _, ok := indexList[lib.Index]; !ok {
			indexList[lib.Index] = factories.NewBleveIndex(
				cfg.GetString("bleve.path"), lib.Index, entities.NewBookIndexMapping(),
			)
		}

		res[lib.Name] = repos.NewBooksInfo(batchSize, false, storage, indexList[lib.Index], logger,
			jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
			jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
			nil, nil, nil,
		)
	}

	return res
}

func GetPool(threads int, bus chan pipeline.Doer, rules ...pipeline.DoerDecorator) pipeline.Dispatcher {
	workers := []pipeline.Doer{}
	for i := 0; i < threads; i++ {
		workers = append(workers, pool.NewWorker(bus, rules...))
	}

	return pool.NewPool(bus, workers...)
}
