package factories

import (
	"path"

	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/handlers"
	"gitlab.com/egnd/bookshelf/pkg/echoext"

	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewEchoServer(cfg *viper.Viper, logger zerolog.Logger,
	booksRepo entities.IBooksRepo, fb2Repo entities.IFB2Repo,
) *echo.Echo {
	server := echo.New()
	server.Debug = cfg.GetBool("server.debug")
	server.HideBanner = true
	server.HidePort = true
	server.Renderer = echoext.NewPongoRenderer(server.Debug, nil, map[string]pongo2.FilterFunction{
		"filesize":  echoext.PongoFilterFileSize,
		"trimspace": echoext.PongoFilterTrimSpace,
	})

	server.Use(echoext.NewZeroLogger(cfg, logger))
	if server.Debug {
		echoext.AddPprofHandlers(server)
	} else {
		server.Use(middleware.Recover())
	}

	server.File("/favicon.ico", path.Join(cfg.GetString("markup.theme_dir"), "assets/favicon.ico"))
	server.Static("/assets", path.Join(cfg.GetString("markup.theme_dir"), "assets"))

	server.GET("/", handlers.SearchHandler(cfg.GetString("markup.theme_dir"), booksRepo))
	server.GET("/authors", handlers.SearchAuthorsHandler(cfg.GetString("markup.theme_dir"), booksRepo))
	server.GET("/sequences", handlers.SearchSequencesHandler(cfg.GetString("markup.theme_dir"), booksRepo))
	server.GET("/download/:book_name", handlers.DownloadBookHandler(booksRepo, cfg, logger))
	server.GET("/books/:book_id", handlers.BookDetailsHandler(cfg.GetString("markup.theme_dir"), booksRepo, fb2Repo, logger))

	return server
}
