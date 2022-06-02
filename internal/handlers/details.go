package handlers

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-fb2parse"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func DetailsHandler(
	indexRepo entities.IBooksInfoRepo, repoBooks entities.IBooksLibraryRepo, logger zerolog.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var bookInfo entities.BookInfo
		if bookInfo, err = indexRepo.GetBook(c.Param("book_id")); err != nil {
			c.NoContent(http.StatusNotFound)
			return
		}

		switch path.Ext(bookInfo.Src) {
		case ".fb2", ".zip":
			var book fb2parse.FB2File
			if book, err = repoBooks.GetFB2(bookInfo); err != nil {
				c.NoContent(http.StatusInternalServerError)
				return
			}

			bookInfo.ReadDetails(&book)
		default:
			c.NoContent(http.StatusInternalServerError)
			return fmt.Errorf(
				"details handler error: invalid book type %s", path.Ext(bookInfo.Src),
			)
		}

		return c.Render(http.StatusOK, "books-details.html", pongo2.Context{
			"search_form_action": "/",
			"search_placeholder": "Автор, название книги, серии, ISBN и т.д.",
			"title":              strings.Split(bookInfo.Index.Titles, "; ")[0],

			"book": bookInfo,
		})
	}
}
