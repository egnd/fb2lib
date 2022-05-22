package main

import (
	"flag"
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"

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

	cfg := factories.NewViperCfg(*cfgPath, *cfgPrefix)
	logger := factories.NewZerologLogger(cfg, os.Stderr)

	libs, err := entities.NewLibraries("libraries", cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("init libs cfg")
	}

	index := factories.NewCompositeBleveIndex(cfg.GetString("bleve.path"), libs, entities.NewBookIndexMapping())
	defer index.Close()

	storage := factories.NewBoltDB(cfg.GetString("boltdb.path"))
	defer storage.Close()

	repoLibrary := repos.NewLibraryFiles(libs)
	repoInfo := repos.NewBooksInfo(0, cfg.GetBool("bleve.highlight"), storage, index, logger,
		jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
	)

	server, err := factories.NewEchoServer(libs, cfg, logger, repoInfo, repoLibrary)
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
