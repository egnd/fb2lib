package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/egnd/go-wpool/v2"
	"github.com/egnd/go-wpool/v2/interfaces"
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
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "BS", "Prefix for env variables.")

	resetIndex = flag.Bool("reset", false, "Clear indexes first.")
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

	cfg, err := factories.NewViperCfg(*cfgPath, *cfgPrefix)
	if err != nil {
		log.Fatal().Err(err).Msg("init config")
	}

	logger := factories.NewZerologLogger(cfg, os.Stderr)

	if *profiler != "" {
		defer RunProfiler(*profiler, cfg).Stop()
	}

	libs, err := GetLibs(*libName, cfg, logger)
	if err != nil {
		log.Fatal().Err(err).Msg("init config")
	}

	bar, err := GetProgressBar(libs, *hideBar, cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("init log file output")
	}

	pipeline := make(chan interfaces.Worker, *buffSize)
	pool := wpool.NewPipelinePool(pipeline, wpool.NewZerologAdapter(logger))
	defer pool.Close()
	for i := runtime.NumCPU(); i > 0; i-- {
		pool.AddWorker(wpool.NewPipelineWorker(pipeline))
	}

	repoList, err := GetRepos(*resetIndex, *batchSize, libs, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("init repos")
	}

	semaphore := make(chan struct{}, *threadsCnt)
	defer close(semaphore)

	if err = IterateLibs(libs, repoList, &wg, logger,
		pool, bar, &cntTotal, &cntIndexed, semaphore); err != nil {
		logger.Fatal().Err(err).Msg("iterate libs")
	}

	wg.Wait()
	time.Sleep(1 * time.Second)

	for _, repo := range repoList {
		repo.Close()
	}

	logger.Info().
		Uint32("succeed", cntIndexed.Total()).
		Uint32("failed", cntTotal.Total()-cntIndexed.Total()).
		Dur("dur", time.Since(startTS)).Msg("indexing finished")
}

func IterateLibs(libs entities.Libraries, repoList map[string]entities.IBooksIndexRepo,
	wg *sync.WaitGroup, logger zerolog.Logger, pool interfaces.Pool, bar *mpb.Bar,
	cntTotal *entities.CntAtomic32, cntIndexed *entities.CntAtomic32, semaphore chan struct{},
) error {
	libItems, err := libs.GetItems()
	if err != nil {
		return err
	}

	for libItem, libTitle := range libItems {
		lib := libs[libTitle]
		itemPath := strings.TrimPrefix(libItem, lib.BooksDir)
		logger := logger.With().Str("libname", lib.Name).Str("libitem", itemPath).Logger()

		logger.Debug().Msg("lib item found")

		switch path.Ext(libItem) {
		case ".zip":
			wg.Add(1)
			semaphore <- struct{}{}
			go func() {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				tasks.NewZipIndexTask(libItem, lib,
					repoList[lib.IndexDir], logger, bar, cntTotal, cntIndexed, wg, pool,
				).Do()
			}()
		case ".fb2":
			wg.Add(1)
			if err := pool.AddTask(tasks.NewFB2IndexTask(lib.Name, itemPath, libItem,
				repoList[lib.IndexDir], logger, bar, cntTotal, cntIndexed, wg,
			)); err != nil {
				logger.Error().Err(err).Msg("send fb2 to pool")
			}
		default:
			logger.Warn().Str("type", path.Ext(libItem)).Msg("undefined lib item type")
		}
	}

	return nil
}

func GetRepos(reset bool, batchSize int, libs entities.Libraries, logger zerolog.Logger) (map[string]entities.IBooksIndexRepo, error) {
	repoList := map[string]entities.IBooksIndexRepo{}
	for _, lib := range libs {
		if _, ok := repoList[lib.IndexDir]; ok {
			continue
		}

		if reset {
			files, _ := filepath.Glob(path.Join(lib.IndexDir, "*"))
			for _, f := range files {
				if err := os.RemoveAll(f); err != nil && !os.IsNotExist(err) {
					return nil, err
				}
			}
		}

		index, err := factories.NewBleveIndex(lib.IndexDir, entities.NewBookIndexMapping())
		if err != nil {
			return nil, err
		}

		repoList[lib.IndexDir] = repos.NewBooksIndexBleve(batchSize, false, index, logger)
	}

	return repoList, nil
}

func GetLibs(singleLibName string, cfg *viper.Viper, logger zerolog.Logger) (res entities.Libraries, err error) {
	if res, err = entities.NewLibraries("libraries", cfg); err != nil {
		return
	}

	if singleLibName != "" {
		if lib, ok := res[singleLibName]; ok {
			res = entities.Libraries{singleLibName: lib}
		} else {
			err = fmt.Errorf("library %s not found", singleLibName)
		}
	}

	return
}

func GetProgressBar(libs entities.Libraries, disableBar bool, cfg *viper.Viper, logger *zerolog.Logger) (bar *mpb.Bar, err error) {
	if disableBar {
		return
	}

	if err = os.MkdirAll(cfg.GetString("logs.dir"), 0644); err != nil {
		return
	}

	var logOutput *os.File
	if logOutput, err = os.OpenFile(
		fmt.Sprintf("%s/indexing.%d.log", cfg.GetString("logs.dir"), time.Now().Unix()),
		os.O_RDWR|os.O_CREATE, 0644,
	); err != nil {
		return
	}

	*logger = logger.Output(zerolog.ConsoleWriter{Out: logOutput, NoColor: true})

	bar = mpb.New(
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

	return
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
