package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/egnd/go-wpool/v2"
	"github.com/egnd/go-wpool/v2/interfaces"
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

	libItemsThreads = flag.Int("items_threads", 1, "Number of parallel threads for library files reading.")
	readBuffSize    = flag.Int("read_buffer", 0, "Books reading queue size.")
	readThreads     = flag.Int("read_threads", 0, "Number of parallel threads for library books reading (default items_threads*10).")
	parseBuffSize   = flag.Int("parse_buffer", 0, "Books parsing queue size.")
	parseThreads    = flag.Int("parse_threads", 0, "Number of parallel threads for library books parsing (default half of CPU cores count).")
	batchSize       = flag.Int("batchsize", 200, "Books index batch size.")

	cfgPath   = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix = flag.String("env-prefix", "BS", "Prefix for env variables.")
	libName   = flag.String("lib", "", "Handle only specific lib.")
	profiler  = flag.String("pprof", "", "Enable profiler (mem,allocs,heap,cpu,trace,goroutine,mutex,block,thread).")
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

	if *libItemsThreads == 0 {
		*libItemsThreads = 1
	}

	if *readThreads == 0 {
		*readThreads = *libItemsThreads * 10
	}

	if *parseThreads == 0 {
		if *parseThreads = runtime.NumCPU() / 2; *parseThreads == 0 {
			*parseThreads = 1
		}
	}

	if *showVersion {
		fmt.Println(appVersion)
		return
	}

	cfg := factories.NewViperCfg(*cfgPath, *cfgPrefix)
	if *profiler != "" {
		defer RunProfiler(*profiler, cfg).Stop()
	}

	storage := factories.NewBoltDB(cfg.GetString("boltdb.path"))
	defer storage.Close()

	logger := factories.NewZerologLogger(cfg, os.Stderr)
	libs := GetLibs(*libName, cfg, logger)
	if !*hideBar {
		bars = mpb.New(mpb.WithOutput(os.Stdout))
		barTotal = GetProgressBar(bars, libs, cfg, &logger)
	}

	reposInfo := GetInfoRepos(*batchSize, libs, logger, cfg, storage)
	repoMarks := repos.NewLibMarks("indexed_items", storage)

	parsingPipeline := make(chan interfaces.Worker, *parseBuffSize)
	defer close(parsingPipeline)

	parsingPool := GetPool(*readThreads, parsingPipeline, logger)
	defer parsingPool.Close()

	readingPipeline := make(chan interfaces.Worker, *readBuffSize)
	defer close(readingPipeline)

	readingPool := GetPool(*parseThreads, readingPipeline, logger)
	defer readingPool.Close()

	semaphore := make(chan struct{}, *libItemsThreads)
	defer close(semaphore)

	libItems, err := libs.GetItems()
	if err != nil {
		panic(err)
	}

	var num int
	total := len(libItems)
	for _, v := range libItems {
		wg.Add(1)
		num++

		semaphore <- struct{}{}
		go func(num int, libItem, libTitle string) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			tasks.NewHandleLibItemTask(num, total, libItem, libs[libTitle],
				logger, readingPool, parsingPool, &wg, repoMarks, &cntTotal, bars,
				func(data io.Reader, book entities.BookInfo, logger zerolog.Logger) interfaces.Task {
					return tasks.NewIndexFB2DataTask(data, book,
						libs[libTitle].Encoder, reposInfo[libTitle], logger, &wg, &cntIndexed, barTotal,
					)
				},
			).Do()
		}(num, v.Item, v.Lib)
	}

	wg.Wait()

	for _, repo := range reposInfo {
		repo.Close()
	}

	logger.Info().
		Uint32("succeed", cntIndexed.Total()).
		Uint32("failed", cntTotal.Total()-cntIndexed.Total()).
		Dur("dur", time.Since(startTS)).Msg("indexing finished")
}

func GetInfoRepos(batchSize int, libs entities.Libraries, logger zerolog.Logger, cfg *viper.Viper, storage *bbolt.DB) map[string]entities.IBooksInfoRepo {
	res := make(map[string]entities.IBooksInfoRepo, len(libs))

	for _, lib := range libs {
		res[lib.Name] = repos.NewBooksInfo(batchSize, false, storage,
			factories.NewBleveIndex(cfg.GetString("bleve.path"), lib.Index, entities.NewBookIndexMapping()),
			logger,
			jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
			jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
			nil, nil, nil,
		)
	}

	return res
}

func GetLibs(singleLibName string, cfg *viper.Viper, logger zerolog.Logger) entities.Libraries {
	res, err := entities.NewLibraries("libraries", cfg)
	if err != nil {
		panic(err)
	}

	if singleLibName != "" {
		if lib, ok := res[singleLibName]; ok {
			res = entities.Libraries{singleLibName: lib}
		} else {
			panic(fmt.Errorf("library %s not found", singleLibName))
		}
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

func GetPool(workersCNt int, pipeline chan interfaces.Worker, logger zerolog.Logger) interfaces.Pool {
	pool := wpool.NewPipelinePool(pipeline, wpool.NewZerologAdapter(logger))

	for i := 0; i < workersCNt; i++ {
		pool.AddWorker(wpool.NewPipelineWorker(pipeline))
	}

	return pool
}
