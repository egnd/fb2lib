package handlers

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gitlab.com/egnd/bookshelf/internal/repos"
)

func DownloadFB2Handler(repo *repos.BooksBleveRepo, logger zerolog.Logger, cfg *viper.Viper) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		bookID := c.Param("book_id")
		logger = logger.With().Str("book_id", bookID).Str("page", "download").Logger()

		book, err := repo.GetBook(c.Request().Context(), bookID)
		if err != nil {
			logger.Error().Err(err).Msg("get book")
			c.NoContent(http.StatusNotFound)
			return err
		}

		switch path.Ext(book.Src) {
		case ".zip":
			req, err := http.NewRequest(http.MethodGet,
				fmt.Sprintf("http://localhost:%d/library/%s", cfg.GetInt("server.port"), strings.TrimPrefix(book.Src, "web/library/")), nil,
			)
			if err != nil {
				logger.Error().Err(err).Msg("create subreq")
				c.NoContent(http.StatusInternalServerError)
				return err
			}
			req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", book.Offset, book.Offset+book.Size))

			restream, err := http.DefaultClient.Do(req)
			if err != nil {
				logger.Error().Err(err).Msg("subrequest")
				c.NoContent(http.StatusInternalServerError)
				return err
			}
			defer restream.Body.Close()

			c.Response().Header().Set(echo.HeaderContentEncoding, "deflate")
			c.Response().Header().Set(echo.HeaderContentDisposition,
				fmt.Sprintf(`attachment; filename="%s.fb2"`, bookID),
			)

			return c.Stream(http.StatusOK, "application/fb2", restream.Body)
		case ".fb2":
			return c.Attachment(book.Src, bookID+".fb2")
		default:
			err = fmt.Errorf("download book error: invalid src %s", book.Src)
			logger.Error().Err(err).Msg("get book")
			c.NoContent(http.StatusInternalServerError)
			return err
		}
	}
}
