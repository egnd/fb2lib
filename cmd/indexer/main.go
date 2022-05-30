package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
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
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "BS", "Prefix for env variables.")

	hideBar    = flag.Bool("hidebar", false, "Hide progress bar.")
	threadsCnt = flag.Int("threads", 1, "Parallel threads count.")
	buffSize   = flag.Int("bufsize", 0, "Workers pool queue size.")
	batchSize  = flag.Int("batchsize", 100, "Books index batch size.")
	libName    = flag.String("lib", "", "Handle only specific lib.")
	profiler   = flag.String("pprof", "", "Enable profiler (mem,allocs,heap,cpu,trace,goroutine,mutex,block,thread).")
)

func main() {
	var (
		wg         sync.WaitGroup
		cntTotal   entities.CntAtomic32
		cntIndexed entities.CntAtomic32
		startTS    = time.Now()
	)

	flag.Parse()

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
	bar := GetProgressBar(libs, *hideBar, cfg, &logger)
	reposInfo := GetInfoRepos(*batchSize, libs, logger, cfg, storage)
	repoMarks := repos.NewLibMarks("indexed_items", storage)

	pipeline := make(chan interfaces.Worker, *buffSize)
	defer close(pipeline)

	semaphore := make(chan struct{}, *threadsCnt)
	defer close(semaphore)

	pool := GetPool(pipeline, logger)
	defer pool.Close()

	IterateLibs(libs, reposInfo, &wg, logger, pool, bar, &cntTotal, &cntIndexed, semaphore, repoMarks)
	wg.Wait()

	for _, repo := range reposInfo {
		repo.Close()
	}

	logger.Info().
		Uint32("succeed", cntIndexed.Total()).
		Uint32("failed", cntTotal.Total()-cntIndexed.Total()).
		Dur("dur", time.Since(startTS)).Msg("indexing finished")
}

func IterateLibs(libs entities.Libraries, reposInfo map[string]entities.IBooksInfoRepo,
	wg *sync.WaitGroup, logger zerolog.Logger, pool interfaces.Pool, bar *mpb.Bar,
	cntTotal *entities.CntAtomic32, cntIndexed *entities.CntAtomic32, semaphore chan struct{},
	repoMarks entities.ILibMarksRepo,
) {
	libItems, err := libs.GetItems()
	if err != nil {
		panic(err)
	}

	for libItem, libTitle := range libItems {
		lib := libs[libTitle]
		itemPath := strings.TrimPrefix(libItem, lib.Dir)
		logger := logger.With().Str("libname", lib.Name).Str("libitem", itemPath).Logger()

		logger.Debug().Msg("lib item found")

		switch path.Ext(libItem) {
		case ".zip":
			wg.Add(1)
			semaphore <- struct{}{}
			go func(libItem string, lib entities.Library, repoInfo entities.IBooksInfoRepo) {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				tasks.NewZipIndexTask(libItem, lib, repoInfo, logger, bar, cntTotal, cntIndexed, wg, pool, repoMarks).Do()
			}(libItem, lib, reposInfo[libTitle])
		case ".fb2":
			wg.Add(1)
			if err := pool.AddTask(tasks.NewFB2IndexTask(libTitle, lib.Encoder, itemPath, libItem,
				reposInfo[libTitle], logger, bar, cntTotal, cntIndexed, wg, repoMarks,
			)); err != nil {
				logger.Error().Err(err).Msg("send fb2 to pool")
			}
		default:
			logger.Warn().Str("type", path.Ext(libItem)).Msg("undefined lib item type")
		}
	}
}

func GetInfoRepos(batchSize int, libs entities.Libraries, logger zerolog.Logger, cfg *viper.Viper, storage *bbolt.DB) map[string]entities.IBooksInfoRepo {
	res := make(map[string]entities.IBooksInfoRepo, len(libs))

	for libName := range libs {
		res[libName] = repos.NewBooksInfo(batchSize, false, storage,
			factories.NewBleveIndex(cfg.GetString("bleve.path"), libName, entities.NewBookIndexMapping()),
			logger,
			jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
			jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
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

func GetProgressBar(libs entities.Libraries, disableBar bool, cfg *viper.Viper, logger *zerolog.Logger) *mpb.Bar {
	if disableBar {
		return nil
	}

	if err := os.MkdirAll(cfg.GetString("logs.dir"), 0644); err != nil {
		panic(err)
	}

	logOutput, err := os.OpenFile(
		fmt.Sprintf("%s/indexing.%d.log", cfg.GetString("logs.dir"), time.Now().Unix()),
		os.O_RDWR|os.O_CREATE, 0644,
	)
	if err != nil {
		panic(err)
	}

	*logger = logger.Output(zerolog.ConsoleWriter{Out: logOutput, NoColor: true})

	return mpb.New(
		mpb.WithOutput(os.Stdout),
		// mpb.WithWidth(64),
	).New(libs.GetSize(),
		mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),
		mpb.PrependDecorators(
			decor.Name("total - "),
			decor.Elapsed(decor.ET_STYLE_GO),
		),
		mpb.AppendDecorators(
			decor.AverageETA(decor.ET_STYLE_GO),
			decor.Name(" - "),
			decor.AverageSpeed(decor.UnitKB, "% .2f"),
			decor.Name(" - "),
			decor.CountersKibiByte("% .2f/% .2f"),
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

func GetPool(pipeline chan interfaces.Worker, logger zerolog.Logger) interfaces.Pool {
	pool := wpool.NewPipelinePool(pipeline, wpool.NewZerologAdapter(logger))

	for i := runtime.NumCPU(); i > 0; i-- {
		pool.AddWorker(wpool.NewPipelineWorker(pipeline))
	}

	return pool
}
