package factories

import (
	"net/http"
	"path"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/handlers"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/fb2lib/pkg/echoext"
	"github.com/labstack/echo/v4"

	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewEchoServer(version string, libs entities.Libraries, cfg *viper.Viper, logger zerolog.Logger,
	repoInfo *repos.BooksLevelBleve, repoBooks *repos.LibraryFs,
) (*echo.Echo, error) {
	var err error
	server := echo.New()
	server.Debug = cfg.GetBool("server.debug")
	server.HideBanner = true
	server.HidePort = true

	if server.Renderer, err = NewEchoRender(version, cfg, server, repoInfo, logger); err != nil {
		return nil, err
	}

	server.Use(echoext.NewZeroLogger(cfg, logger))
	if server.Debug {
		echoext.AddPprofHandlers(server)
	} else {
		server.Use(middleware.Recover())
	}

	server.File("/favicon.ico", path.Join(cfg.GetString("renderer.dir"), "assets/favicon.ico"))
	server.Static("/assets", path.Join(cfg.GetString("renderer.dir"), "assets"))
	server.GET("/live", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	server.GET("/", func(c echo.Context) error { return c.Redirect(http.StatusMovedPermanently, "/books/") })
	server.GET("/books/", handlers.BooksHandler(cfg, libs, repoInfo, repoBooks, logger))
	server.GET("/books/:tag/:tag_value/", handlers.BooksHandler(cfg, libs, repoInfo, repoBooks, logger))
	server.GET("/download/:book", handlers.DownloadHandler(libs, repoInfo, cfg, logger))
	server.GET("/book/:id", handlers.BookDetailsHandler(repoInfo, repoBooks))
	server.GET("/book/:id/remove", handlers.RemoveBookHandler(repoInfo))
	server.GET("/genres/", handlers.GenresHandler(cfg, repoInfo))
	server.GET("/series/", handlers.SeriesHandler(cfg, repoInfo))
	server.GET("/series/:letter/", handlers.SeriesHandler(cfg, repoInfo))
	server.GET("/authors/", handlers.AuthorsHandler(cfg, repoInfo, repoBooks))
	server.GET("/authors/:letter/", handlers.AuthorsHandler(cfg, repoInfo, repoBooks))
	server.GET("/authors/:letter/:name", handlers.AuthorsHandler(cfg, repoInfo, repoBooks))

	return server, nil
}

func NewEchoRender(version string, cfg *viper.Viper,
	server *echo.Echo, repo *repos.BooksLevelBleve, logger zerolog.Logger,
) (echo.Renderer, error) {
	books, _ := repo.GetCnt(repos.BucketBooks)
	genres, _ := repo.GetCnt(repos.BucketGenres)
	authors, _ := repo.GetCnt(repos.BucketAuthors)
	series, _ := repo.GetCnt(repos.BucketSeries)

	globals := cfg.GetStringMap("renderer.globals")
	globals["sidebar_stats"] = map[string]uint64{
		"books":   books,
		"authors": authors,
		"genres":  genres,
		"series":  series,
	}
	globals["libslist"], _ = repo.GetLibs()
	globals["langslist"], _ = repo.GetLangs()
	globals["app_version"] = version
	globals["debug"] = server.Debug

	return echoext.NewPongoRenderer(echoext.PongoRendererCfg{
		Debug:   server.Debug,
		TplsDir: cfg.GetString("renderer.dir"),
	}, globals, map[string]pongo2.FilterFunction{
		"filesize":  echoext.PongoFilterFileSize,
		"trimspace": echoext.PongoFilterTrimSpace,
	})
}
