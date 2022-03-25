package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func EchoLogger(cfg *viper.Viper, logger zerolog.Logger) echo.MiddlewareFunc {
	pretty := cfg.GetBool("app.pretty")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			start := time.Now()

			event := logger.Info().
				Int("status", c.Response().Status).
				Str("remote_ip", c.RealIP()).
				Str("host", c.Request().Host).
				Str("uri", c.Request().RequestURI).
				Str("method", c.Request().Method)

			if err = next(c); err != nil {
				c.Error(err)
				event = event.Err(err)
			}

			id := c.Request().Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = c.Response().Header().Get(echo.HeaderXRequestID)
			}
			if id != "" {
				event = event.Str("id", id)
			}

			stop := time.Now()

			if pretty {
				event = event.Str("latency", stop.Sub(start).String())
			} else {
				event = event.Dur("latency", stop.Sub(start))
			}

			event.Msg("request")

			return
		}
	}
}
