package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/rs/zerolog/log"

	"gitlab.com/egnd/bookshelf/internal/factories"
	"gitlab.com/egnd/bookshelf/internal/indexing"
	"gitlab.com/egnd/bookshelf/internal/repos"
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

	server := factories.NewEchoServer(cfg, logger,
		repos.NewBooksBleve(cfg.GetBool("bleve.highlight"), booksIndex, logger),
	)
	logger.Info().
		Int("port", cfg.GetInt("server.port")).
		Str("version", appVersion).
		Msg("server is starting...")

	if err = server.Start(fmt.Sprintf(":%d", cfg.GetInt("server.port"))); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}

	logger.Info().Msg("server stopped")
}
