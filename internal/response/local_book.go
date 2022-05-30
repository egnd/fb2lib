package response

import (
	"errors"
	"net/http"
	"path"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
)

func BookAttachment(book entities.BookInfo, libs entities.Libraries, server echo.Context) error {
	bookPath := book.Src
	if lib, ok := libs[book.LibName]; ok {
		bookPath = path.Join(lib.Dir, bookPath)
	} else {
		server.NoContent(http.StatusInternalServerError)
		return errors.New("can't define book library")
	}

	return server.Attachment(bookPath, entities.BuildBookName(book.Index)+path.Ext(book.Src))
}
