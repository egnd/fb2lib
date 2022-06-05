package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

func AuthorsHandler(repo entities.IBooksInfoRepo) echo.HandlerFunc {
	return func(c echo.Context) error {
		letter, err := url.QueryUnescape(c.Param("letter"))
		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return err
		}

		name, err := url.QueryUnescape(c.Param("name"))
		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return err
		}

		var title string
		var books []entities.BookInfo
		var series map[string]int

		authors, err := repo.GetAuthorsByLetter(letter)
		if err != nil {
			c.NoContent(http.StatusInternalServerError)
			return err
		}

		if name != "" {
			title = "Автор " + name

			books, err = repo.GetOtherAuthorBooks(name, nil)
			if err != nil {
				c.NoContent(http.StatusInternalServerError)
				return err
			}

			series, err = repo.GetOtherAuthorSeries(name, "")
			if err != nil {
				c.NoContent(http.StatusInternalServerError)
				return err
			}

		} else {
			title = "Авторы"
			if letter != "" {
				title += fmt.Sprintf(` на букву "%s"`, letter)
			}
		}

		return c.Render(http.StatusOK, "pages/authors.html", pongo2.Context{
			"section_name": "authors",
			"page_title":   title,
			"page_h1":      title,

			"cur_letter": letter,
			"cur_name":   name,
			"authors":    authors,
			"series":     series,
			"books":      books,
		})
	}
}
