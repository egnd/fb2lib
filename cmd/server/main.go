package main

import (
	"flag"
	"fmt"

	"os"

	"github.com/rs/zerolog/log"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/factories"
	"github.com/egnd/fb2lib/internal/repos"
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

	libs, err := entities.NewLibraries("libraries", cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("init libs cfg")
	}

	booksIndex, err := factories.NewCompositeBleveIndex(libs, entities.NewBookIndexMapping())
	if err != nil {
		logger.Fatal().Err(err).Msg("init index")
	}

	repoIndex := repos.NewBooksIndexBleve(cfg.GetBool("bleve.highlight"), booksIndex, logger)
	repoBooks := repos.NewBooksDataFiles(libs)
	server, err := factories.NewEchoServer(libs, cfg, logger, repoIndex, repoBooks)
	if err != nil {
		logger.Fatal().Err(err).Msg("init http server")
	}

	logger.Info().
		Int("port", cfg.GetInt("server.port")).
		Str("version", appVersion).
		Msg("server is starting...")

	if err = server.Start(fmt.Sprintf(":%d", cfg.GetInt("server.port"))); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}

	logger.Info().Msg("server stopped")
}
