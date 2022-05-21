package handlers

import (
	"fmt"
	"net/http"
	"path"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func DetailsHandler(
	indexRepo entities.IBooksIndexRepo, repoBooks entities.IBooksDataRepo, logger zerolog.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var bookIdx entities.BookIndex
		if bookIdx, err = indexRepo.GetBook(c.Param("book_id")); err != nil {
			c.NoContent(http.StatusNotFound)
			return
		}

		var book entities.FB2Book

		switch path.Ext(bookIdx.Src) {
		case ".fb2", ".zip":
			if book, err = repoBooks.GetFB2(bookIdx); err != nil {
				c.NoContent(http.StatusInternalServerError)
				return
			}
		default:
			c.NoContent(http.StatusInternalServerError)
			return fmt.Errorf(
				"details handler error: invalid book type %s", path.Ext(bookIdx.Src),
			)
		}

		return c.Render(http.StatusOK, "books-details.html", pongo2.Context{
			"search_form_action": "/",
			"search_placeholder": "Автор, название книги, серии, ISBN и т.д.",
			"title":              book.Description.TitleInfo.BookTitle,

			"book":     book,
			"book_idx": bookIdx,
		})
	}
}
