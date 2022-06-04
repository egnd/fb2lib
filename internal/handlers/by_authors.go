package handlers

// import (
// 	"net/http"

// 	"github.com/egnd/fb2lib/internal/entities"
// 	"github.com/egnd/fb2lib/pkg/pagination"
// 	"github.com/flosch/pongo2/v5"
// 	"github.com/labstack/echo/v4"
// )

// func ByAuthorsHandler(genresLimit int,
// 	repoInfo entities.IBooksInfoRepo,
// 	repoBooks entities.IBooksLibraryRepo,
// ) echo.HandlerFunc {
// 	return func(c echo.Context) (err error) {
// 		searchQuery := c.QueryParam("q")

// 		pager := pagination.NewPager(c.Request()).SetPageSize(20).
// 			ReadPageSize().ReadCurPage()

// 		var books []entities.BookInfo
// 		books, err = repoInfo.SearchByAuthor(searchQuery, pager)

// 		if err != nil {
// 			c.NoContent(http.StatusBadRequest)
// 			return
// 		}

// 		addDetails(books, repoBooks)

// 		var genresShort entities.GenresIndex
// 		if genresShort, err = repoInfo.GetGenres(genresLimit); err != nil {
// 			c.NoContent(http.StatusBadRequest)
// 			return
// 		}

// 		return c.Render(http.StatusOK, "books-list.html", pongo2.Context{
// 			"search_query":       searchQuery,
// 			"search_placeholder": "ФИО автора",
// 			"search_type":        "authors",
// 			"title":              "Поиск по авторам",
// 			"genres_short":       genresShort,

// 			"books": books,
// 			"pager": pager,
// 		})
// 	}
// }
