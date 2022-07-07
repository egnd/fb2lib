package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

func AuthorsHandler(
	repoInfo *repos.BooksBadgerBleve,
	repoBooks *repos.LibraryFs,
) echo.HandlerFunc {
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
		var books []entities.Book
		var series entities.FreqsItems
		var breadcrumbs entities.BreadCrumbs

		authors, err := repoInfo.GetAuthorsByPrefix(letter)
		if err != nil {
			c.NoContent(http.StatusInternalServerError)
			return err
		}

		if name != "" {
			title = "Автор " + name

			books, err = repoInfo.GetAuthorsBooks(10, []string{name}, nil)
			if err != nil {
				c.NoContent(http.StatusInternalServerError)
				return err
			}

			repoBooks.AppendFB2Books(books)

			series, err = repoInfo.GetAuthorsSeries([]string{name}, nil)
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

		if letter == "" {
			breadcrumbs = breadcrumbs.Push("Авторы", "")
		} else {
			breadcrumbs = breadcrumbs.Push("Авторы", "/authors/")
			if name == "" {
				breadcrumbs = breadcrumbs.Push(letter, "")
			} else {
				breadcrumbs = breadcrumbs.Push(letter, "/authors/"+letter+"/").Push(name, "")
			}
		}

		return c.Render(http.StatusOK, "pages/authors.html", pongo2.Context{
			"section_name": "authors",
			"page_title":   title,
			"page_h1":      title,

			"cur_letter":  letter,
			"cur_name":    name,
			"authors":     authors,
			"series":      series,
			"books":       books,
			"breadcrumbs": breadcrumbs,
		})
	}
}
