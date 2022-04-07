package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	wpool "github.com/egnd/go-wpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vbauerster/mpb/v7"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/factories"
	"gitlab.com/egnd/bookshelf/internal/tasks"
	"gitlab.com/egnd/bookshelf/pkg/library"
)

var (
	appVersion = "debug"

	showVersion = flag.Bool("version", false, "Show app version.")
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "BS", "Prefix for env variables.")

	rewriteIndex = flag.Bool("rewrite", false, "Rewrite existing indexes.")
	workersCnt   = flag.Int("workers", 1, "Index workers count.")
	bufSize      = flag.Int("bufsize", 0, "Workers pool queue buffer size.")
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

	logger := factories.NewZerologLogger(cfg, os.Stderr)
	ctx := context.Background()

	pool := wpool.NewPool(wpool.PoolCfg{
		WorkersCnt:   uint(*workersCnt),
		TasksBufSize: uint(*bufSize),
	}, func(num uint, pipeline chan wpool.IWorker) wpool.IWorker {
		return wpool.NewWorker(ctx, wpool.WorkerCfg{Pipeline: pipeline})
	})
	defer pool.Close()

	var wg sync.WaitGroup
	var cntTotal entities.CntAtomic32
	var cntIndexed entities.CntAtomic32

	startTS := time.Now()
	bar := mpb.New(
		mpb.WithOutput(os.Stderr),
	)

	if err = library.NewLocalFSItems(
		cfg.GetString("extractor.dir"), []string{".zip"}, logger,
	).IterateItems(func(libFile os.FileInfo, libDir string, num, total int, logger zerolog.Logger) error {
		wg.Add(1)
		return pool.Add(tasks.NewBooksArchiveIndexTask(
			libFile, libDir, cfg.GetString("bleve.books_dir"),
			*rewriteIndex, &cntTotal, &cntIndexed,
			logger, &wg, bar,
		))
	}); err != nil {
		logger.Error().Err(err).Msg("handle library items")
	}

	wg.Wait()
	time.Sleep(time.Second)

	logger.Info().
		Uint32("total", cntIndexed.Total()).
		Uint32("indexed", cntIndexed.Total()).
		Dur("dur", time.Now().Sub(startTS)).
		Msg("indexing finished")
}
