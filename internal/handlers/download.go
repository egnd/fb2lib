package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/response"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func DownloadHandler(libs entities.Libraries,
	repo entities.IBooksInfoRepo, cfg *viper.Viper, logger zerolog.Logger,
) echo.HandlerFunc {
	converterDir := cfg.GetString("converter.dir")
	if err := os.MkdirAll(converterDir, 0755); err != nil {
		panic(err)
	}

	return func(c echo.Context) (err error) {
		var bookID string
		bookType := strings.Trim(path.Ext(c.Param("book_id")), ".")
		switch bookType {
		case "epub", "fb2":
			bookID = strings.TrimSuffix(c.Param("book_id"), "."+bookType)
		default:
			c.NoContent(http.StatusBadRequest)
			return
		}

		var book entities.BookInfo
		if book, err = repo.GetBook(bookID); err != nil {
			c.NoContent(http.StatusNotFound)
			return
		}

		switch {
		case bookType == "fb2" && path.Ext(book.Src) == ".fb2" && strings.Contains(book.Src, ".zip"):
			err = response.FB2FromLocalZip(book, libs, c)
		case bookType == "fb2" && path.Ext(book.Src) == ".fb2":
			err = response.BookAttachment(book, libs, c)
		case bookType == "epub" && path.Ext(book.Src) == ".fb2" && strings.Contains(book.Src, ".zip"):
			err = response.ConvertFB2EpubZipOffset(converterDir, book, libs, c, logger)
		case bookType == "epub" && path.Ext(book.Src) == ".fb2":
			err = response.ConvertFB2Epub(converterDir, book, libs, c, logger)
		default:
			err = fmt.Errorf("download %s book error: invalid src %s", bookType, book.Src)
		}

		if err != nil {
			c.NoContent(http.StatusInternalServerError)
		}

		return
	}
}
