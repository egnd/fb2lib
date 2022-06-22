package handlers

import (
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
)

func SeriesHandler(repo entities.IBooksInfoRepo) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// letter := c.Param("letter")

		// series, err := repo.GetSeriesByLetter(letter)
		// if err != nil {
		// 	c.NoContent(http.StatusInternalServerError)
		// 	return
		// }

		// title := "Книжные серии"
		// if letter != "" {
		// 	title += fmt.Sprintf(`, начинающиеся с буквы "%s"`, letter)
		// }

		// return c.Render(http.StatusOK, "pages/series.html", pongo2.Context{
		// 	"section_name": "series",
		// 	"page_title":   title,
		// 	"page_h1":      title,

		// 	"cur_letter": letter,
		// 	"series":     series,
		// })
		return nil
	}
}
