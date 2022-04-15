package echoext

import (
	"net/http"
	"net/http/pprof"

	"github.com/labstack/echo/v4"
)

func AddPprofHandlers(server *echo.Echo) {
	server.GET("/debug/pprof/", NewPProfHandler(pprof.Index))
	server.GET("/debug/pprof/allocs", NewPProfHandler(pprof.Handler("allocs").ServeHTTP))
	server.GET("/debug/pprof/block", NewPProfHandler(pprof.Handler("block").ServeHTTP))
	server.GET("/debug/pprof/cmdline", NewPProfHandler(pprof.Cmdline))
	server.GET("/debug/pprof/goroutine", NewPProfHandler(pprof.Handler("goroutine").ServeHTTP))
	server.GET("/debug/pprof/heap", NewPProfHandler(pprof.Handler("heap").ServeHTTP))
	server.GET("/debug/pprof/mutex", NewPProfHandler(pprof.Handler("mutex").ServeHTTP))
	server.GET("/debug/pprof/profile", NewPProfHandler(pprof.Profile))
	server.GET("/debug/pprof/threadcreate", NewPProfHandler(pprof.Handler("threadcreate").ServeHTTP))
	server.GET("/debug/pprof/trace", NewPProfHandler(pprof.Trace))
}

func NewPProfHandler(handler http.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		handler(ctx.Response().Writer, ctx.Request())

		return nil
	}
}
