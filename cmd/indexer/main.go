package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	wpool "github.com/egnd/go-wpool"
	"github.com/pkg/profile"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/factories"
	"github.com/egnd/fb2lib/internal/tasks"
	"github.com/egnd/fb2lib/pkg/library"
)

var (
	appVersion = "debug"

	showVersion = flag.Bool("version", false, "Show app version.")
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "BS", "Prefix for env variables.")

	resetIndex = flag.Bool("reset", false, "Reset indexe.")
	hideBar    = flag.Bool("hidebar", false, "Hide progress bar.")
	workersCnt = flag.Int("workers", 1, "Index workers count.")
	buffSize   = flag.Int("bufsize", 0, "Workers pool queue buffer size.")
	batchSize  = flag.Int("batchsize", 100, "Books index batch size.")
	libName    = flag.String("lib", "", "Handle only specific lib.")
	profiler   = flag.String("pprof", "", "Enable profiler (mem,allocs,heap,cpu,trace,goroutine,mutex,block,thread).")

	wg         sync.WaitGroup
	cntTotal   entities.CntAtomic32
	cntIndexed entities.CntAtomic32
	bar        *mpb.Progress
	totalBar   *mpb.Bar

	libFormats = []string{".zip"} // @TODO: tar.gz
	ctx        = context.Background()
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println(appVersion)
		return
	}

	cfg, err := factories.NewViperCfg(*cfgPath, *cfgPrefix)
	if err != nil {
		log.Fatal().Err(err).Msg("init config")
	}

	if *profiler != "" {
		defer profilerStart(*profiler, cfg).Stop()
	}

	logger := factories.NewZerologLogger(cfg, os.Stderr)
	startTS := time.Now()

	pool := newPool(*workersCnt, *buffSize)
	defer pool.Close()

	libs, err := entities.NewCfgLibsMap(cfg, *libName)
	if err != nil {
		logger.Fatal().Err(err).Msg("init config")
	}

	if !*hideBar {
		bar = mpb.New(
			mpb.WithOutput(os.Stdout),
		)

		logOutput, err := newLogFileOutput(cfg)
		if err != nil {
			logger.Fatal().Err(err).Msg("init log file output")
		}
		defer logOutput.Close()

		logger = logger.Output(zerolog.ConsoleWriter{Out: logOutput, NoColor: true})

		if totalBar = getTotalBar(libFormats, libs, bar); totalBar == nil {
			logger.Fatal().Msg("unable to init total bar")
		}
	}

	var wg sync.WaitGroup
	for libName, lib := range libs {
		logger := logger.With().Str("libname", libName).Logger()

		if err := processLib(lib, logger, pool, &wg); err != nil {
			logger.Fatal().Err(err).Msg("handle library")
		}
	}

	wg.Wait()
	time.Sleep(1 * time.Second)

	logger.Info().Dur("dur", time.Since(startTS)).
		Uint32("total", cntIndexed.Total()).
		Uint32("indexed", cntIndexed.Total()).
		Msg("indexing finished")
}

func getTotalBar(targetFormats []string, libs entities.CfgLibsMap, bar *mpb.Progress) *mpb.Bar {
	var totalSize int64

	for _, lib := range libs {
		library.NewLocalFSItems(lib.BooksDir, targetFormats, zerolog.Nop()).
			IterateItems(func(libFile os.FileInfo, libDir string, num, total int, logger zerolog.Logger) error {
				totalSize += libFile.Size()
				return nil
			})
	}

	if totalSize == 0 {
		return nil
	}

	return bar.AddBar(totalSize,
		mpb.PrependDecorators(
			decor.Name("total"),
		),
		mpb.AppendDecorators(
			decor.Elapsed(decor.ET_STYLE_GO),
			decor.Name(" - "),
			decor.CountersKibiByte("% .2f/% .2f"),
		),
	)
}

func profilerStart(profType string, cfg *viper.Viper) interface{ Stop() } {
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

func newPool(workersCnt, bufsize int) *wpool.Pool {
	return wpool.NewPool(wpool.PoolCfg{
		WorkersCnt:   uint(workersCnt),
		TasksBufSize: uint(bufsize),
	}, func(num uint, pipeline chan wpool.IWorker) wpool.IWorker {
		return wpool.NewWorker(ctx, wpool.WorkerCfg{Pipeline: pipeline})
	})
}

func newLogFileOutput(cfg *viper.Viper) (out *os.File, err error) {
	if err = os.MkdirAll(cfg.GetString("logs.dir"), 0644); err != nil {
		return
	}

	return os.OpenFile(
		fmt.Sprintf("%s/indexing.%d.log", cfg.GetString("logs.dir"), time.Now().Unix()),
		os.O_RDWR|os.O_CREATE, 0644,
	)
}

func processLib(
	lib entities.CfgLibrary,
	logger zerolog.Logger, pool *wpool.Pool, wg *sync.WaitGroup,
) (err error) {
	var index bleve.Index
	if _, statErr := os.Stat(lib.IndexDir); *resetIndex || statErr != nil {
		os.RemoveAll(lib.IndexDir)
		os.MkdirAll(lib.IndexDir, 0644)
		index, err = bleve.New(lib.IndexDir, entities.NewBookIndexMapping())
	} else {
		index, err = bleve.Open(lib.IndexDir)
	}

	if err != nil {
		return
	}

	return library.NewLocalFSItems(
		lib.BooksDir, libFormats, logger,
	).IterateItems(func(libFile os.FileInfo, libDir string, num, total int, logger zerolog.Logger) error {
		switch path.Ext(libFile.Name()) {
		case ".zip":
			wg.Add(1)
			return pool.Add(tasks.NewZIPFB2IndexTask(
				libFile, libDir, lib.IndexDir, *batchSize, &cntTotal, &cntIndexed, logger, wg, bar, totalBar, index,
			))
		default:
			return nil
		}
	})
}
