package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/response"
)

func DownloadEpubHandler(repo entities.IBooksRepo, cfg *viper.Viper, logger zerolog.Logger) echo.HandlerFunc {
	converterDir := cfg.GetString("library.converter_dir")
	if err := os.MkdirAll(converterDir, 0755); err != nil {
		panic(err)
	}

	return func(c echo.Context) (err error) {
		var book entities.BookIndex
		if book, err = repo.GetBook(c.Request().Context(), c.Param("book_id")); err != nil {
			c.NoContent(http.StatusNotFound)

			return
		}

		switch {
		case path.Ext(book.Src) == ".zip" || (path.Ext(book.Src) == ".fb2" && strings.Contains(book.Src, ".zip")):
			err = response.ConvertZipFB2Epub(converterDir, book, c, logger)
		case path.Ext(book.Src) == ".epub":
			err = c.Attachment(book.Src, book.ID+".epub")
		default:
			err = fmt.Errorf("download epub book error: invalid book src %s", book.Src)
		}

		if err != nil {
			c.NoContent(http.StatusInternalServerError)
		}

		return
	}
}
