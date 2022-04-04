package factories

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/handlers"
	"gitlab.com/egnd/bookshelf/internal/middleware"
	"gitlab.com/egnd/bookshelf/internal/repos"
	"gitlab.com/egnd/bookshelf/pkg/library"

	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	pongo2echo "github.com/stnc/pongo2echo"
)

func NewEchoServer(cfg *viper.Viper, logger zerolog.Logger,
	booksRepo *repos.BooksBleveRepo, extractor library.IExtractorFactory,
) *echo.Echo {
	server := echo.New()
	server.Debug = cfg.GetBool("debug")
	server.HideBanner = true
	server.HidePort = true
	server.Renderer = pongo2echo.Renderer{Debug: server.Debug}

	server.Use(middleware.EchoLogger(cfg, logger))
	if !server.Debug {
		server.Use(echomiddleware.Recover())
	}

	server.File("/favicon.ico", cfg.GetString("markup.theme_dir")+"/favicon.ico")
	server.Static("/markup", cfg.GetString("markup.theme_dir"))

	if cfg.GetString("extractor.uri_prefix") != "" {
		server.Static(cfg.GetString("extractor.uri_prefix"), cfg.GetString("extractor.dir"))
	}

	server.GET("/live", handlers.EchoLiveHandler())
	server.GET("/", handlers.MainPageHandler(booksRepo, logger))
	server.GET("/download/:book_id/fb2", handlers.DownloadFB2Handler(booksRepo, logger, cfg, extractor))
	// server.GET("/download/:book_id/epub", handlers.MainPageHandler()) // @TODO: https://github.com/rupor-github/fb2converter
	// server.GET("/authors/", handlers.MainPageHandler()) // @TODO:
	// server.GET("/authors/:id", handlers.MainPageHandler()) // @TODO:
	// server.GET("/sequences/", handlers.MainPageHandler()) // @TODO:
	// server.GET("/sequences/:id", handlers.MainPageHandler()) // @TODO:
	// server.GET("/genres/", handlers.MainPageHandler()) // @TODO:
	// server.GET("/genres/:id", handlers.MainPageHandler()) // @TODO:

	if server.Debug {
		server.Static("/html", cfg.GetString("markup.html_dir"))
	}

	return server
}
