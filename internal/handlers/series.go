package handlers

import (
	"fmt"
	"net/http"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func SeriesHandler(cfg *viper.Viper, repo *repos.BooksLevelBleve) echo.HandlerFunc {
	defPageSize := cfg.GetInt("renderer.globals.series_size")

	return func(c echo.Context) (err error) {
		letter := c.Param("letter")
		pager := pagination.NewPager(c.Request()).SetPageSize(defPageSize).ReadPageSize().ReadCurPage()

		series, err := repo.GetSeriesByPrefix(letter, pager)
		if err != nil {
			c.NoContent(http.StatusInternalServerError)
			return
		}

		var breadcrumbs entities.BreadCrumbs
		title := "Книжные серии"
		if letter != "" {
			title += fmt.Sprintf(`, начинающиеся с "%s"`, letter)
			breadcrumbs = breadcrumbs.Push("Cерии", "/series/").Push(letter, "")
		} else {
			breadcrumbs = breadcrumbs.Push("Cерии", "")
		}

		return c.Render(http.StatusOK, "pages/series.html", pongo2.Context{
			"section_name": "series",
			"page_title":   title,
			"page_h1":      title,

			"cur_letter":  letter,
			"series":      series,
			"breadcrumbs": breadcrumbs,
			"pager":       pager,
		})
	}
}
