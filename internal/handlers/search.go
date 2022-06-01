package handlers

import (
	"net/http"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/egnd/go-fb2parse"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

func SearchHandler(
	repoInfo entities.IBooksInfoRepo,
	repoBooks entities.IBooksLibraryRepo,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		searchQuery := c.QueryParam("q")

		pager := pagination.NewPager(c.Request()).SetPageSize(20).
			ReadPageSize().ReadCurPage()

		var books []entities.BookInfo
		books, err = repoInfo.SearchAll(searchQuery, pager)

		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		addDetails(books, repoBooks)

		return c.Render(http.StatusOK, "books-list.html", pongo2.Context{
			"search_query":       searchQuery,
			"search_placeholder": "Автор, название книги, серии, ISBN и т.д.",
			"search_type":        "all",

			"books": books,
			"pager": pager,
		})
	}
}

func addDetails(books []entities.BookInfo, repo entities.IBooksLibraryRepo) (err error) {
	var book fb2parse.FB2File
	var wg sync.WaitGroup

	for k, info := range books {
		wg.Add(1)
		go func(k int, info entities.BookInfo) {
			defer wg.Done()
			if book, err = repo.GetFB2(info); err != nil {
				return
			}

			books[k].ReadDetails(&book)

			if len(books[k].Details.Images) > 0 {
				books[k].Details.Images = books[k].Details.Images[0:1]
			}
		}(k, info)
	}

	wg.Wait()

	return
}
