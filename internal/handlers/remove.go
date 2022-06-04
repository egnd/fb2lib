package handlers

import (
	"net/http"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
)

func RemoveBookHandler(repo entities.IBooksInfoRepo) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		if err = repo.Remove(c.Param("id")); err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		if strings.Contains(c.Request().Referer(), "/book/") {
			return c.Redirect(http.StatusMovedPermanently, "/")
		} else {
			return c.Redirect(http.StatusMovedPermanently, c.Request().Referer())
		}
	}
}
