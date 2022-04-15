package response

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

func LocalFile(ext string, book entities.BookIndex, server echo.Context) error {
	return server.Attachment(book.Src, BuildBookName(book)+ext)
}
