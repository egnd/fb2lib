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
	"github.com/vbauerster/mpb/v7/decor"

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

	rewriteIndex    = flag.Bool("rewrite", false, "Rewrite existing indexes.")
	extendedMapping = flag.Bool("extmapping", false, "Use extended index mapping.")
	hideBar         = flag.Bool("hidebar", false, "Hide progress bar.")
	workersCnt      = flag.Int("workers", 1, "Index workers count.")
	bufSize         = flag.Int("bufsize", 0, "Workers pool queue buffer size.")

	wg         sync.WaitGroup
	cntTotal   entities.CntAtomic32
	cntIndexed entities.CntAtomic32
	bar        *mpb.Progress
	totalBar   *mpb.Bar

	targetLibFormats = []string{".zip"}
	ctx              = context.Background()
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
	startTS := time.Now()

	pool := wpool.NewPool(wpool.PoolCfg{
		WorkersCnt:   uint(*workersCnt),
		TasksBufSize: uint(*bufSize),
	}, func(num uint, pipeline chan wpool.IWorker) wpool.IWorker {
		return wpool.NewWorker(ctx, wpool.WorkerCfg{Pipeline: pipeline})
	})
	defer pool.Close()

	if !*hideBar {
		bar = mpb.New(
			mpb.WithOutput(os.Stdout),
		)

		os.Remove("var/indexing.log")
		logOutput, err := os.OpenFile("var/indexing.log", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			logger.Fatal().Err(err).Msg("init log file output")
		}
		defer logOutput.Close()

		logger = logger.Output(zerolog.ConsoleWriter{Out: logOutput, NoColor: true})

		if totalBar = getTotalBar(targetLibFormats, cfg.GetString("extractor.dir"), bar); totalBar == nil {
			logger.Fatal().Msg("unable to init total bar")
		}
	}

	if err = library.NewLocalFSItems(
		cfg.GetString("extractor.dir"), targetLibFormats, logger,
	).IterateItems(func(libFile os.FileInfo, libDir string, num, total int, logger zerolog.Logger) error {
		wg.Add(1)
		return pool.Add(tasks.NewBooksArchiveIndexTask(
			libFile, libDir, cfg.GetString("bleve.books_dir"),
			*rewriteIndex, *extendedMapping, &cntTotal, &cntIndexed,
			logger, &wg, bar, totalBar,
		))
	}); err != nil {
		logger.Error().Err(err).Msg("handle library items")
	}

	wg.Wait()
	time.Sleep(2 * time.Second)

	logger.Info().
		Uint32("total", cntIndexed.Total()).
		Uint32("indexed", cntIndexed.Total()).
		Dur("dur", time.Now().Sub(startTS)).
		Msg("indexing finished")
}

func getTotalBar(targetFormats []string, libDir string, bar *mpb.Progress) *mpb.Bar {
	var totalSize int64
	library.NewLocalFSItems(libDir, targetFormats, zerolog.Nop()).
		IterateItems(func(libFile os.FileInfo, libDir string, num, total int, logger zerolog.Logger) error {
			totalSize += libFile.Size()
			return nil
		})

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
			decor.CountersKiloByte("% .2f/% .2f"),
		),
	)
}
