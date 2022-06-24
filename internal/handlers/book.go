package handlers

import (
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/labstack/echo/v4"
)

func BookDetailsHandler(repo *repos.BooksBadgerBleve) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// var book entities.BookInfo
		// if book, err = repo.FindByID(c.Param("id")); err != nil {
		// 	c.NoContent(http.StatusNotFound)
		// 	return
		// }

		// var otherSeriesBooks, otherAuthorBooks []entities.BookInfo
		// var otherSeries map[string]int

		// if otherSeriesBooks, err = repo.GetSeriesBooks(book.Index.Serie, &book); err != nil {
		// 	c.NoContent(http.StatusInternalServerError)
		// 	return
		// }

		// if otherAuthorBooks, err = repo.GetOtherAuthorBooks(book.Index.Author, &book); err != nil {
		// 	c.NoContent(http.StatusInternalServerError)
		// 	return
		// }

		// if otherSeries, err = repo.GetOtherAuthorSeries(book.Index.Author, book.Index.Serie); err != nil {
		// 	c.NoContent(http.StatusInternalServerError)
		// 	return
		// }

		// return c.Render(http.StatusOK, "pages/book.html", pongo2.Context{
		// 	"page_title":         "Книга " + strings.Split(book.Index.Title, entities.IndexFieldSep)[0],
		// 	"page_h1":            "Книга " + strings.Split(book.Index.Title, entities.IndexFieldSep)[0],
		// 	"book":               book,
		// 	"other_series_books": otherSeriesBooks,
		// 	"other_series":       otherSeries,
		// 	"other_books":        otherAuthorBooks,
		// })
		return nil
	}
}
