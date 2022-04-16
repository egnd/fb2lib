package handlers

import (
	"net/http"
	"path"

	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

func BookDetailsHandler(tplsDir string,
	indexRepo entities.IBooksRepo, fb2Repo entities.IFB2Repo, logger zerolog.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var bookIdx entities.BookIndex
		if bookIdx, err = indexRepo.GetBook(c.Param("book_id")); err != nil {
			c.NoContent(http.StatusNotFound)
			return
		}

		var book entities.FB2Book
		if book, err = fb2Repo.GetFor(bookIdx); err != nil {
			c.NoContent(http.StatusInternalServerError)
			return
		}

		return c.Render(http.StatusOK, path.Join(tplsDir, "books-details.html"), pongo2.Context{
			"search_form_action": "/",
			"search_placeholder": "Автор, название книги, серии, ISBN и т.д.",

			"book":     book,
			"book_idx": bookIdx,
		})
	}
}
