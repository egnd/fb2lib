package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func BooksHandler(cfg *viper.Viper, libs entities.Libraries,
	repoInfo entities.IBooksInfoRepo,
	repoBooks entities.IBooksLibraryRepo,
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
		pager := pagination.NewPager(c.Request()).SetPageSize(defPageSize).
			ReadPageSize().ReadCurPage()

		var books []entities.BookInfo
		if books, err = repoInfo.FindBooks(searchQuery, tag, tagValue, pager); err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		var title string
		switch tag {
		case "author":
			title = fmt.Sprintf(`Поиск в книгах автора "%s"`, tagValue)
		case "transl":
			title = fmt.Sprintf(`Поиск книг в переводе "%s"`, tagValue)
		case "serie":
			title = fmt.Sprintf(`Поиск в книгах серии "%s"`, tagValue)
		case "genre":
			title = fmt.Sprintf(`Поиск книг в жанре "%s"`, tagValue)
		case "publ":
			title = fmt.Sprintf(`Поиск в книгах издателя "%s"`, tagValue)
		case "lang":
			title = fmt.Sprintf(`Поиск книг на языке %s`, tagValue)
		case "lib":
			title = fmt.Sprintf(`Поиск книг в коллекции "%s"`, tagValue)
		default:
			title = "Поиск по книгам"
		}

		return c.Render(http.StatusOK, "pages/books.html", pongo2.Context{
			"section_name": "books",
			"page_title":   title,
			"page_h1":      title,

			"search_query": searchQuery,
			"cur_tag":      tag,
			"cur_tag_val":  tagValue,
			"books":        books,
			"pager":        pager,
			"libs": func() (res []string) {
				for lib := range libs {
					res = append(res, lib)
				}
				return
			}(),
		})
	}
}
