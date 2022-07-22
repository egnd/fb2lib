package handlers

import (
	"net/http"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

func BookDetailsHandler(
	repoBooks *repos.BooksLevelBleve,
	repoLib *repos.LibraryFs,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var book *entities.Book
		if book, err = repoBooks.GetByID(c.Param("id")); err != nil {
			c.NoContent(http.StatusNotFound)
			return
		}

		if err = repoLib.AppendFB2Book(book); err != nil {
			c.NoContent(http.StatusInternalServerError)
			return
		}

		var seriesBooks, authorsBooks []entities.Book
		var series entities.FreqsItems

		if seriesBooks, err = repoBooks.GetSeriesBooks(100, book.Series(), book); err != nil {
			c.NoContent(http.StatusInternalServerError)
			return
		}

		repoLib.AppendFB2Books(seriesBooks)

		if authorsBooks, err = repoBooks.GetAuthorsBooks(100, book.Authors(), book); err != nil {
			c.NoContent(http.StatusInternalServerError)
			return
		}

		repoLib.AppendFB2Books(authorsBooks)

		if series, err = repoBooks.GetAuthorsSeries(book.Authors(), book.Series()); err != nil {
			c.NoContent(http.StatusInternalServerError)
			return
		}

		return c.Render(http.StatusOK, "pages/book.html", pongo2.Context{
			"page_title":     "Книга " + book.Info.Title,
			"page_h1":        book.Info.Title,
			"book":           book,
			"series_books":   seriesBooks,
			"authors_books":  authorsBooks,
			"authors_series": series,
			"breadcrumbs":    (entities.BreadCrumbs{}).Push("Книги", "/books/").Push(book.Info.Title, ""),
		})
	}
}
