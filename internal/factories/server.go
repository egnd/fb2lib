package factories

import (
	"net/http"
	"path"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/handlers"
	"github.com/egnd/fb2lib/pkg/echoext"
	"github.com/labstack/echo/v4"

	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewEchoServer(libs entities.Libraries, cfg *viper.Viper, logger zerolog.Logger,
	repoInfo entities.IBooksInfoRepo, repoBooks entities.IBooksLibraryRepo,
) (*echo.Echo, error) {
	var err error
	server := echo.New()
	server.Debug = cfg.GetBool("server.debug")
	server.HideBanner = true
	server.HidePort = true

	if server.Renderer, err = NewEchoRender(cfg, server); err != nil {
		return nil, err
	}

	server.Use(echoext.NewZeroLogger(cfg, logger))
	if server.Debug {
		echoext.AddPprofHandlers(server)
	} else {
		server.Use(middleware.Recover())
	}

	server.File("/favicon.ico", path.Join(cfg.GetString("renderer.tpl_dir"), "assets/favicon.ico"))
	server.Static("/assets", path.Join(cfg.GetString("renderer.tpl_dir"), "assets"))
	server.GET("/live", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	server.GET("/", handlers.SearchHandler(cfg, libs, repoInfo, repoBooks))
	server.GET("/lib/:lib_name", handlers.SearchHandler(cfg, libs, repoInfo, repoBooks))
	server.GET("/download/:book", handlers.DownloadHandler(libs, repoInfo, cfg, logger))

	// server.GET("/book/:id", handlers.SearchHandler())
	// server.GET("/book/:id/remove", handlers.SearchHandler())

	// server.GET("/genres/", handlers.SearchHandler())
	// server.GET("/genres/:name", handlers.SearchHandler())
	// server.GET("/authors/", handlers.SearchHandler())
	// server.GET("/authors/:name", handlers.SearchHandler())
	// server.GET("/series/", handlers.SearchHandler())
	// server.GET("/series/:name", handlers.SearchHandler())

	// server.GET("/", handlers.SearchHandler(cfg.GetInt("renderer.sidebar.genres_size"), repoInfo, repoBooks))
	// server.GET("/genres/", handlers.GenresHandler(cfg.GetInt("renderer.sidebar.genres_size"), repoInfo, repoBooks))
	// server.GET("/by_authors/", handlers.ByAuthorsHandler(cfg.GetInt("renderer.sidebar.genres_size"), repoInfo, repoBooks))
	// server.GET("/by_series/", handlers.BySeriesHandler(cfg.GetInt("renderer.sidebar.genres_size"), repoInfo, repoBooks))
	// server.GET("/details/:book_id", handlers.DetailsHandler(cfg.GetInt("renderer.sidebar.genres_size"), repoInfo, repoBooks, logger))

	return server, nil
}

func NewEchoRender(cfg *viper.Viper, server *echo.Echo) (echo.Renderer, error) {
	return echoext.NewPongoRenderer(echoext.PongoRendererCfg{
		Debug:   server.Debug,
		TplsDir: cfg.GetString("renderer.tpl_dir"),
	}, cfg.GetStringMap("renderer.globals"), map[string]pongo2.FilterFunction{
		"filesize":  echoext.PongoFilterFileSize,
		"trimspace": echoext.PongoFilterTrimSpace,
	})
}
