package handlers

import (
	"errors"
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

func DownloadBookHandler(libsCfg entities.CfgLibsMap,
	repo entities.IBooksIndexRepo, cfg *viper.Viper, logger zerolog.Logger,
) echo.HandlerFunc {
	converterDir := cfg.GetString("converter.dir")
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

		if lib, ok := libsCfg[book.LibName]; ok {
			book.Src = path.Join(lib.BooksDir, book.Src)
		} else {
			c.NoContent(http.StatusInternalServerError)
			return errors.New("can't define book library")
		}

		switch {
		case bookType == "fb2":
			err = response.FB2FromLocalZip(book, c)
		case bookType == "epub" && path.Ext(book.Src) == ".zip":
			err = response.ConvertFB2EpubZipOffset(converterDir, book, c, logger)
		case bookType == "epub" && path.Ext(book.Src) == ".fb2":
			err = response.ConvertFB2Epub(converterDir, book, c, logger)
		default:
			err = fmt.Errorf("download %s book error: invalid src %s", bookType, book.Src)
		}

		if err != nil {
			c.NoContent(http.StatusInternalServerError)
		}

		return
	}
}
