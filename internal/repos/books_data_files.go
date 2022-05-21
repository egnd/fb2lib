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

type BooksDataFiles struct {
	libs entities.Libraries
}

func NewBooksDataFiles(libs entities.Libraries) *BooksDataFiles {
	return &BooksDataFiles{
		libs: libs,
	}
}

func (r *BooksDataFiles) GetFB2(book entities.BookIndex) (res entities.FB2Book, err error) {
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
			strings.Split(path.Join(lib.BooksDir, book.Src), ".zip")[0]+".zip",
			int64(book.Offset), int64(book.SizeCompressed),
		)
	}

	return r.extractFB2(path.Join(lib.BooksDir, book.Src))
}

func (r *BooksDataFiles) extractZippedFB2(archivePath string, offset int64, limit int64) (res entities.FB2Book, err error) {
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

func (r *BooksDataFiles) extractFB2(filePath string) (res entities.FB2Book, err error) {
	var fb2Stream io.ReadCloser
	if fb2Stream, err = os.Open(filePath); err != nil {
		return
	}
	defer fb2Stream.Close()

	err = fb2parser.UnmarshalStream(fb2Stream, &res)

	return
}
