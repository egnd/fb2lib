package handlers

import (
	"net/http"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/flosch/pongo2"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/repos"
)

func MainPageHandler(repo *repos.BooksBleveRepo, logger zerolog.Logger) echo.HandlerFunc {
	pageSize := 20

	return func(c echo.Context) (err error) {
		searchQuery := c.QueryParam("q")
		pager := pagination.NewPaginator(c.Request(), pageSize, 0)

		var books []entities.BookIndex
		books, err = repo.GetBooks(c.Request().Context(), searchQuery, pager)

		if err != nil {
			logger.Error().Err(err).Str("query", searchQuery).Str("page", "main").Msg("get books")
			c.NoContent(http.StatusBadRequest)
			return
		}

		return c.Render(http.StatusOK, "web/tpls/main.html", pongo2.Context{
			"h1":                 "Домашняя библиотека",
			"search_action":      "/",
			"search_query":       searchQuery,
			"search_placeholder": "Автор, название книги, серии, ISBN и т.д.",

			"books": books,
			"pager": pager,
		})
	}
}
