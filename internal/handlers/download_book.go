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

func DownloadBookHandler(repo entities.IBooksIndexRepo, cfg *viper.Viper, logger zerolog.Logger) echo.HandlerFunc {
	converterDir := cfg.GetString("library.converter_dir")
	if err := os.MkdirAll(converterDir, 0755); err != nil {
		panic(err)
	}

	return func(c echo.Context) (err error) {
		var bookID string
		bookType := strings.Trim(path.Ext(c.Param("book_name")), ".")
		switch bookType {
		case "epub", "fb2":
			bookID = strings.TrimSuffix(c.Param("book_name"), "."+bookType)
		default:
			c.NoContent(http.StatusNotFound)
			return
		}

		var book entities.BookIndex
		if book, err = repo.GetBook(bookID); err != nil {
			c.NoContent(http.StatusNotFound)
			return
		}

		switch {
		case bookType == "fb2" && path.Ext(book.Src) == ".zip":
			err = response.FB2FromLocalZip(book, c)
		case bookType == "fb2" && path.Ext(book.Src) == ".fb2" && strings.Contains(book.Src, ".zip"+string(os.PathSeparator)):
			err = response.FB2FromLocalZip(book, c)
		case bookType == "fb2" && path.Ext(book.Src) == ".fb2" && !strings.Contains(book.Src, ".zip"+string(os.PathSeparator)):
			err = response.LocalFile(".fb2", book, c)
		case bookType == "epub" && path.Ext(book.Src) == ".zip":
			err = response.ConvertFB2EpubZipOffset(converterDir, book, c, logger)
		case bookType == "epub" && path.Ext(book.Src) == ".fb2":
			err = response.ConvertFB2Epub(converterDir, book, c, logger)
		case bookType == "epub" && path.Ext(book.Src) == ".epub" && !strings.Contains(book.Src, ".zip"+string(os.PathSeparator)):
			err = response.LocalFile(".epub", book, c)
		// case  bookType == "epub" && path.Ext(book.Src) == ".epub" && strings.Contains(book.Src, ".zip"+string(os.PathSeparator)): // @TODO:
		default:
			err = fmt.Errorf("download %s book error: invalid src %s", bookType, book.Src)
		}

		if err != nil {
			c.NoContent(http.StatusInternalServerError)
		}

		return
	}
}
