package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/response"
)

func DownloadFB2Handler(repo entities.IBooksRepo, cfg *viper.Viper) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var book entities.BookIndex
		if book, err = repo.GetBook(c.Param("book_id")); err != nil {
			c.NoContent(http.StatusNotFound)

			return
		}

		switch {
		case path.Ext(book.Src) == ".zip" || strings.Contains(book.Src, ".zip"+string(os.PathSeparator)):
			err = response.FB2FromLocalZip(book, c)
		case path.Ext(book.Src) == ".fb2":
			err = response.LocalFile(".fb2", book, c)
		default:
			err = fmt.Errorf("download fb2 book error: invalid book src %s", book.Src)
		}

		if err != nil {
			c.NoContent(http.StatusInternalServerError)
		}

		return
	}
}
