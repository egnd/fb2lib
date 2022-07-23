package response

import (
	"errors"
	"net/http"
	"path"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
)

func BookAttachment(book *entities.Book, libs entities.Libraries, server echo.Context) error {
	bookPath := book.Src
	if lib, ok := libs[book.Lib]; ok {
		bookPath = path.Join(lib.Dir, bookPath)
	} else {
		server.NoContent(http.StatusInternalServerError)
		return errors.New("can't define book library")
	}

	return server.Attachment(bookPath, entities.BuildBookName(book)+path.Ext(book.Src))
}
