package response

import (
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
)

func FB2FromLocalZip(book *entities.Book, libs entities.Libraries, server echo.Context) error {
	zipFilePath := strings.Split(book.Src, ".zip")[0] + ".zip"
	if lib, ok := libs[book.Lib]; ok {
		zipFilePath = path.Join(lib.Dir, zipFilePath)
	} else {
		server.NoContent(http.StatusInternalServerError)
		return errors.New("can't define book library")
	}

	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	reader := flate.NewReader(io.NewSectionReader(zipFile, int64(book.Offset), int64(book.SizeCompressed)))
	defer reader.Close()

	server.Response().Header().Set(echo.HeaderContentDisposition,
		fmt.Sprintf(`attachment; filename="%s.fb2"`, entities.BuildBookName(book)),
	)

	return server.Stream(http.StatusOK, "application/fb2", reader)
}
