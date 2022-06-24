package handlers

import (
	"net/http"

	"github.com/egnd/fb2lib/internal/repos"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func GenresHandler(cfg *viper.Viper, repo *repos.BooksBadgerBleve) echo.HandlerFunc {
	defPageSize := cfg.GetInt("renderer.globals.genres_size")

	return func(c echo.Context) (err error) {
		pager := pagination.NewPager(c.Request()).SetPageSize(defPageSize).ReadPageSize().ReadCurPage()

		genres, err := repo.GetGenres(pager)
		if err != nil {
			c.NoContent(http.StatusBadRequest)
			return
		}

		return c.Render(http.StatusOK, "pages/genres.html", pongo2.Context{
			"section_name": "genres",
			"page_title":   "Список жанров",
			"page_h1":      "Список жанров",

			"genres": genres,
			"pager":  pager,
		})
	}
}
