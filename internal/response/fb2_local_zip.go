package response

import (
	"compress/flate"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

func FB2FromLocalZip(book entities.BookIndex, server echo.Context) error {
	zipFile, err := os.Open(strings.Split(book.Src, ".zip")[0] + ".zip")
	if err != nil {
		return err
	}
	defer zipFile.Close()

	reader := flate.NewReader(io.NewSectionReader(zipFile, int64(book.Offset), int64(book.SizeCompressed)))
	defer reader.Close()

	server.Response().Header().Set(echo.HeaderContentDisposition,
		fmt.Sprintf(`attachment; filename="%s.fb2"`, book.ID),
	)

	return server.Stream(http.StatusOK, "application/fb2", reader)
}