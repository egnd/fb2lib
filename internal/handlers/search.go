package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func SearchHandler(cfg *viper.Viper, libs entities.Libraries,
	repoInfo entities.IBooksInfoRepo,
	repoBooks entities.IBooksLibraryRepo,
) echo.HandlerFunc {
	defPageSize, err := strconv.Atoi(strings.Split(cfg.GetString("renderer.globals.limits_books"), ",")[0])
	if err != nil {
		panic(err)
	}

	return func(c echo.Context) (err error) {
		libName := c.Param("lib_name")
		if _, ok := libs[libName]; libName != "" && !ok {
			c.NoContent(http.StatusNotFound)
			return
		}

		searchQuery := c.QueryParam("q")
		pager := pagination.NewPager(c.Request()).SetPageSize(defPageSize).
			ReadPageSize().ReadCurPage()

		var books []entities.BookInfo
		if books, err = repoInfo.FindIn(libName, searchQuery, pager); err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		libsNames := make([]string, 0, len(libs))
		for lib := range libs {
			libsNames = append(libsNames, lib)
		}

		return c.Render(http.StatusOK, "pages/search.html", pongo2.Context{
			"page_h1":      "Поиск в библиотеке " + libName,
			"search_query": searchQuery,
			"cur_lib":      libName,
			"libs":         libsNames,
			"books":        books,
			"pager":        pager,
		})
	}
}
