package handlers

import (
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/pkg/library"
)

func DownloadFB2Handler(
	repo entities.IBooksRepo, logger zerolog.Logger, cfg *viper.Viper, extractor library.IExtractorFactory,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		bookID := c.Param("book_id")
		logger = logger.With().Str("book_id", bookID).Str("page", "download").Logger()

		book, err := repo.GetBook(c.Request().Context(), bookID)
		if err != nil {
			logger.Error().Err(err).Msg("get book")
			return c.NoContent(http.StatusNotFound)
		}

		switch path.Ext(book.Src) {
		case ".zip":
			extr, err := extractor(book.Src)
			if err != nil {
				logger.Error().Err(err).Msg("init extractor")
				return c.NoContent(http.StatusInternalServerError)
			}

			defer extr.Close()

			stream, err := extr.GetSection(int64(book.Offset), int64(book.Size))
			if err != nil {
				logger.Error().Err(err).Msg("extract book")
				return c.NoContent(http.StatusNotFound)
			}
			defer stream.Close()

			c.Response().Header().Set(echo.HeaderContentEncoding, "deflate")
			c.Response().Header().Set(echo.HeaderContentDisposition,
				fmt.Sprintf(`attachment; filename="%s.fb2"`, bookID),
			)

			return c.Stream(http.StatusOK, "application/fb2", stream)
		case ".fb2":
			return c.Attachment(book.Src, bookID+".fb2")
		default:
			err = fmt.Errorf("download book error: invalid src %s", book.Src)
			logger.Error().Err(err).Msg("get book")
			return c.NoContent(http.StatusInternalServerError)
		}
	}
}
