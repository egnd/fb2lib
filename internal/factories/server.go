package factories

import (
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

func NewEchoServer(cfg *viper.Viper, logger zerolog.Logger,
	booksRepo entities.IBooksIndexRepo, fb2Repo entities.IBooksDataRepo,
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

	server.GET("/", handlers.SearchHandler(booksRepo))
	server.GET("/authors", handlers.SearchAuthorsHandler(booksRepo))
	server.GET("/sequences", handlers.SearchSequencesHandler(booksRepo))
	server.GET("/download/:book_name", handlers.DownloadBookHandler(booksRepo, cfg, logger))
	server.GET("/books/:book_id", handlers.BookDetailsHandler(booksRepo, fb2Repo, logger))

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
