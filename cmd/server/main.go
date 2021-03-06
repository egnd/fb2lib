package main

import (
	"flag"
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/factories"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/go-pipeline/pools"
)

var (
	appVersion = "debug"

	showVersion = flag.Bool("version", false, "Show app version.")
	cfgPath     = flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix   = flag.String("env-prefix", "FBL", "Prefix for env variables.")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println(appVersion)
		return
	}

	cfg := factories.NewViperCfg(*cfgPath, *cfgPrefix)
	logger := factories.NewZerolog(cfg, os.Stderr)

	libs, err := entities.NewLibraries("libraries", cfg)
	if err != nil {
		panic(err)
	}

	repoLibrary := repos.NewLibraryFs(libs, pools.NewSemaphore(20, nil), logger)
	repoBooks := repos.NewBooksLevelBleve(0,
		map[repos.BucketType]*leveldb.DB{
			repos.BucketBooks:   factories.NewLevelDB(cfg.GetString("adapters.leveldb.dir"), "books"),
			repos.BucketAuthors: factories.NewLevelDB(cfg.GetString("adapters.leveldb.dir"), "authors"),
			repos.BucketSeries:  factories.NewLevelDB(cfg.GetString("adapters.leveldb.dir"), "series"),
			repos.BucketGenres:  factories.NewLevelDB(cfg.GetString("adapters.leveldb.dir"), "genres"),
			repos.BucketLibs:    factories.NewLevelDB(cfg.GetString("adapters.leveldb.dir"), "libs"),
			repos.BucketLangs:   factories.NewLevelDB(cfg.GetString("adapters.leveldb.dir"), "langs"),
		},
		factories.NewBleveIndex(cfg.GetString("adapters.bleve.dir"), "books", entities.NewBookIndexMapping()),
		jsoniter.ConfigCompatibleWithStandardLibrary.Marshal,
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal,
		logger,
	)
	defer repoBooks.Close()

	server, err := factories.NewEchoServer(appVersion, libs, cfg, logger, repoBooks, repoLibrary)
	if err != nil {
		logger.Fatal().Err(err).Msg("init http server")
	}

	// logger.Info().Msg("warmup cache...")
	// if err := cacheWarmup(logger, cfg, repoBooks); err != nil {
	// 	panic(err)
	// }

	logger.Info().
		Int("port", cfg.GetInt("server.port")).
		Str("version", appVersion).
		Msg("server is listening...")

	if err = server.Start(fmt.Sprintf(":%d", cfg.GetInt("server.port"))); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}

	logger.Info().Msg("server stopped")
}

// func cacheWarmup(logger zerolog.Logger, cfg *viper.Viper,
// 	repoInfo entities.IBooksInfoRepo,
// ) error {
// 	defPageSize, err := strconv.Atoi(strings.Split(cfg.GetString("renderer.globals.books_sizes"), ",")[0])
// 	if err != nil {
// 		return err
// 	}

// 	if _, err := repoInfo.FindBooks("", "", "", pagination.NewPager(nil).SetPageSize(defPageSize)); err != nil {
// 		return err
// 	}

// 	if _, err := repoInfo.GetStats(); err != nil {
// 		return err
// 	}

// 	if _, err := repoInfo.GetGenres(nil); err != nil {
// 		return err
// 	}

// 	return nil
// }
