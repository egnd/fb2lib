package handlers

import (
	"net/http"

	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/pkg/pagination"
)

func SearchAuthorsHandler(repo entities.IBooksIndexRepo) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		searchQuery := c.QueryParam("q")

		pager := pagination.NewPager(c.Request()).SetPageSize(20).
			ReadPageSize().ReadCurPage()

		var books []entities.BookIndex
		books, err = repo.SearchByAuthor(searchQuery, pager)

		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		return c.Render(http.StatusOK, "books-list.html", pongo2.Context{
			"search_query":       searchQuery,
			"search_placeholder": "ФИО автора",
			"search_type":        "authors",

			"books": books,
			"pager": pager,
		})
	}
}
