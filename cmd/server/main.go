package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/factories"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/fb2lib/pkg/pagination"
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
		panic(err)
	}

	index := factories.NewCompositeBleveIndex(cfg.GetString("bleve.path"), libs, entities.NewBookIndexMapping())
	defer index.Close()

	storage := factories.NewBoltDB(cfg.GetString("boltdb.path"))
	defer storage.Close()

	repoLibrary := repos.NewLibraryFiles(libs)
	repoInfo := repos.NewBooksInfo(0, cfg.GetBool("bleve.highlight"), storage, index, logger,
		jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
		cache.New(time.Hour, 30*time.Minute), repoLibrary,
	)

	server, err := factories.NewEchoServer(libs, cfg, logger, repoInfo, repoLibrary)
	if err != nil {
		logger.Fatal().Err(err).Msg("init http server")
	}

	logger.Info().Msg("warmup cache...")
	if err := cacheWarmup(logger, cfg, repoInfo); err != nil {
		panic(err)
	}

	logger.Info().
		Int("port", cfg.GetInt("server.port")).
		Str("version", appVersion).
		Msg("server is listening...")

	if err = server.Start(fmt.Sprintf(":%d", cfg.GetInt("server.port"))); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}

	logger.Info().Msg("server stopped")
}

func cacheWarmup(logger zerolog.Logger, cfg *viper.Viper,
	repoInfo entities.IBooksInfoRepo,
) error {
	defPageSize, err := strconv.Atoi(strings.Split(cfg.GetString("renderer.globals.limits_books"), ",")[0])
	if err != nil {
		return err
	}

	if _, err := repoInfo.FindIn("", "", pagination.NewPager(nil).SetPageSize(defPageSize)); err != nil {
		return err
	}

	return nil
}
