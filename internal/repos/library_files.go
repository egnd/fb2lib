package repos

import (
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/fb2parser"
)

type LibraryFiles struct {
	libs entities.Libraries
}

func NewLibraryFiles(libs entities.Libraries) *LibraryFiles {
	return &LibraryFiles{
		libs: libs,
	}
}

func (r *LibraryFiles) GetFB2(book entities.BookInfo) (res entities.FB2Book, err error) {
	if book.Src == "" {
		err = errors.New("repo getfb2 error: empty src")
		return
	}

	lib, libExist := r.libs[book.LibName]
	if !libExist {
		err = fmt.Errorf("repo getfb2 error: undefined book lib %s", book.LibName)
		return
	}

	if strings.Contains(book.Src, ".zip") {
		return r.extractZippedFB2(
			strings.Split(path.Join(lib.Dir, book.Src), ".zip")[0]+".zip",
			int64(book.Offset), int64(book.SizeCompressed),
		)
	}

	return r.extractFB2(path.Join(lib.Dir, book.Src))
}

func (r *LibraryFiles) extractZippedFB2(archivePath string, offset int64, limit int64) (res entities.FB2Book, err error) {
	var zipFile *os.File
	if zipFile, err = os.Open(archivePath); err != nil {
		return
	}
	defer zipFile.Close()

	fb2Stream := flate.NewReader(io.NewSectionReader(zipFile, offset, limit))
	defer fb2Stream.Close()

	err = fb2parser.UnmarshalStream(fb2Stream, &res)

	return
}

func (r *LibraryFiles) extractFB2(filePath string) (res entities.FB2Book, err error) {
	var fb2Stream io.ReadCloser
	if fb2Stream, err = os.Open(filePath); err != nil {
		return
	}
	defer fb2Stream.Close()

	err = fb2parser.UnmarshalStream(fb2Stream, &res)

	return
}
