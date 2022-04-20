package response

import (
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
)

func LocalFile(ext string, book entities.BookIndex, server echo.Context) error {
	return server.Attachment(book.Src, BuildBookName(book)+ext)
}
