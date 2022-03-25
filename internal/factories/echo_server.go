package factories

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/handlers"
	"gitlab.com/egnd/bookshelf/internal/middleware"

	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	pongo2echo "github.com/stnc/pongo2echo"
)

func NewEchoServer(cfg *viper.Viper, logger zerolog.Logger,
	booksRepo entities.IBooksRepo,
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

	server.GET("/live", handlers.EchoLiveHandler())

	return server
}
