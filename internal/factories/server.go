package factories

import (
	"path"

	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/handlers"
	"gitlab.com/egnd/bookshelf/internal/middleware"
	"gitlab.com/egnd/bookshelf/pkg/pprof2echo"
	"gitlab.com/egnd/bookshelf/pkg/render2echo"

	"github.com/flosch/pongo2/v5"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewEchoServer(cfg *viper.Viper, logger zerolog.Logger,
	booksRepo entities.IBooksRepo,
) *echo.Echo {
	server := echo.New()
	server.Debug = cfg.GetBool("server.debug")
	server.HideBanner = true
	server.HidePort = true
	server.Renderer = render2echo.NewPongoRenderer(server.Debug, nil, map[string]pongo2.FilterFunction{
		"filesize": render2echo.FilterFileSize,
	})

	server.Use(middleware.EchoLogger(cfg, logger))
	if server.Debug {
		pprof2echo.AddHandlersTo(server)
	} else {
		server.Use(echomiddleware.Recover())
	}

	server.File("/favicon.ico", path.Join(cfg.GetString("markup.theme_dir"), "assets/favicon.ico"))
	server.Static("/assets", path.Join(cfg.GetString("markup.theme_dir"), "assets"))
	server.GET("/", handlers.MainPageHandler(cfg.GetString("markup.theme_dir"), booksRepo, logger))
	server.GET("/download/:book_id/fb2", handlers.DownloadFB2Handler(booksRepo, cfg))
	server.GET("/download/:book_id/epub", handlers.DownloadEpubHandler(booksRepo, cfg, logger))

	return server
}
