package handlers

// import (
// 	"net/http"

// 	"github.com/egnd/fb2lib/internal/entities"
// 	"github.com/flosch/pongo2/v5"
// 	"github.com/labstack/echo/v4"
// )

// func GenresHandler(genresLimit int,
// 	repoInfo entities.IBooksInfoRepo,
// 	repoBooks entities.IBooksLibraryRepo,
// ) echo.HandlerFunc {
// 	return func(c echo.Context) (err error) {
// 		searchQuery := c.QueryParam("q")

// 		var genresShort entities.GenresIndex
// 		if genresShort, err = repoInfo.GetGenres(genresLimit); err != nil {
// 			c.NoContent(http.StatusBadRequest)
// 			return
// 		}

// 		var genres entities.GenresIndex
// 		if genres, err = repoInfo.GetGenres(0); err != nil {
// 			c.NoContent(http.StatusBadRequest)
// 			return
// 		}

// 		return c.Render(http.StatusOK, "books-genres.html", pongo2.Context{
// 			"search_form_action": "/",
// 			"search_query":       searchQuery,
// 			"search_type":        "genres",
// 			"genres_short":       genresShort,

// 			"genres": genres,
// 		})
// 	}
// }
