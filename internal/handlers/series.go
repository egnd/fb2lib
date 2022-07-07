package handlers

import (
	"fmt"
	"net/http"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/repos"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

func SeriesHandler(repo *repos.BooksBadgerBleve) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		letter := c.Param("letter")

		series, err := repo.GetSeriesByPrefix(letter)
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
		})
	}
}
