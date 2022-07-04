package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func BooksHandler(cfg *viper.Viper, libs entities.Libraries,
	repoInfo *repos.BooksBadgerBleve,
	repoBooks *repos.LibraryFs,
	logger zerolog.Logger,
) echo.HandlerFunc {
	defPageSize, err := strconv.Atoi(strings.Split(cfg.GetString("renderer.globals.books_sizes"), ",")[0])
	if err != nil {
		panic(err)
	}

	return func(c echo.Context) (err error) {
		tag, err := url.PathUnescape(c.Param("tag"))
		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		tagValue, err := url.QueryUnescape(c.Param("tag_value"))
		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		searchQuery := c.QueryParam("q")
		pager := pagination.NewPager(c.Request()).SetPageSize(defPageSize).ReadPageSize().ReadCurPage()
		title := "Поиск по книгам"

		var breadcrumbs entities.BreadCrumbs
		if tagValue != "" {
			breadcrumbs = breadcrumbs.Push("Книги", "/books/").Push(tagValue, "")
		} else {
			breadcrumbs = breadcrumbs.Push("Книги", "")
		}

		switch entities.IndexField(tag) {
		case entities.IdxFAuthor:
			title += fmt.Sprintf(` автора "%s"`, tagValue)
		case entities.IdxFTranslator:
			title += fmt.Sprintf(` в переводе "%s"`, tagValue)
		case entities.IdxFSerie:
			title += fmt.Sprintf(` серии "%s"`, tagValue)
		case entities.IdxFGenre:
			title += fmt.Sprintf(` в жанре "%s"`, tagValue)
		case entities.IdxFPublisher:
			title += fmt.Sprintf(` издателя "%s"`, tagValue)
		case entities.IdxFLang:
			title += fmt.Sprintf(` на языке %s`, tagValue)
		case entities.IdxFLib:
			title += fmt.Sprintf(` в коллекции "%s"`, tagValue)
		}

		var books []entities.Book
		if books, err = repoInfo.FindBooks(searchQuery, entities.IndexField(tag), tagValue, pager); err != nil {
			c.NoContent(http.StatusInternalServerError)
			return
		}

		repoBooks.AppendFB2Books(books)

		return c.Render(http.StatusOK, "pages/books.html", pongo2.Context{
			"section_name": "books",
			"page_title":   title,
			"page_h1":      title,

			"search_query": searchQuery,
			"cur_tag":      tag,
			"cur_tag_val":  tagValue,
			"books":        books,
			"pager":        pager,
			"breadcrumbs":  breadcrumbs,
			"libs": func() (res []string) {
				for _, lib := range libs {
					if !lib.Disabled {
						res = append(res, lib.Name)
					}
				}
				return
			}(),
		})
	}
}
