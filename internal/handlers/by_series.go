package handlers

import (
	"net/http"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

func BySeriesHandler(repo entities.IBooksInfoRepo) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		searchQuery := c.QueryParam("q")

		pager := pagination.NewPager(c.Request()).SetPageSize(20).
			ReadPageSize().ReadCurPage()

		var books []entities.BookInfo
		books, err = repo.SearchBySequence(searchQuery, pager)

		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		return c.Render(http.StatusOK, "books-list.html", pongo2.Context{
			"search_query":       searchQuery,
			"search_placeholder": "Название серии книг",
			"search_type":        "sequences",
			"title":              "Поиск по книжным сериям",

			"books": books,
			"pager": pager,
		})
	}
}
