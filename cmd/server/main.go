package main

import (
	"flag"
	"fmt"
	"net/http"
	"path"

	"os"

	"github.com/rs/zerolog/log"

	"gitlab.com/egnd/bookshelf/internal/factories"
	"gitlab.com/egnd/bookshelf/internal/indexing"
	"gitlab.com/egnd/bookshelf/internal/repos"
	"gitlab.com/egnd/bookshelf/pkg/library"
)

var (
	appVersion = "debug"

	showVersion = flag.Bool("version", false, "Show app version.")
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "BS", "Prefix for env variables.")
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

	booksIndex, err := indexing.OpenIndex(cfg.GetString("bleve.index_dir"), logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("init index")
	}

	var extractor library.IExtractorFactory
	if cfg.GetString("extractor.type") == "local" {
		extractor = library.FactoryZipExtractorLocal() // @TODO: make working
	} else {
		extractor = library.FactoryZipExtractorHTTP(
			"http://localhost:"+path.Join(cfg.GetString("server.port"), cfg.GetString("extractor.uri_prefix")),
			cfg.GetString("extractor.dir"), http.DefaultClient,
		)
	}

	server := factories.NewEchoServer(cfg, logger,
		repos.NewBooksBleve(cfg.GetBool("bleve.highlight"), booksIndex, logger),
		extractor,
	)
	logger.Info().Int("port", cfg.GetInt("server.port")).Msg("server is starting...")

	if err = server.Start(fmt.Sprintf(":%d", cfg.GetInt("server.port"))); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}

	logger.Info().Msg("server stopped")
}
