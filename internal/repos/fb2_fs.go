package repos

import (
	"compress/flate"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"

	"gitlab.com/egnd/bookshelf/internal/entities"
	"golang.org/x/net/html/charset"
)

type FB2FilesRepo struct {
}

func NewFB2FilesRepo() *FB2FilesRepo {
	return &FB2FilesRepo{}
}

func (r *FB2FilesRepo) GetFor(book entities.BookIndex) (res entities.FB2Book, err error) {
	if book.Src == "" {
		err = errors.New("fb2repo GetFor error: empty src")
		return
	}

	var fb2Stream io.ReadCloser

	if strings.Contains(book.Src, ".zip") {
		var zipFile *os.File
		if zipFile, err = os.Open(strings.Split(book.Src, ".zip")[0] + ".zip"); err != nil {
			return
		}
		defer zipFile.Close()

		fb2Stream = flate.NewReader(io.NewSectionReader(zipFile, int64(book.Offset), int64(book.SizeCompressed)))
		defer fb2Stream.Close()
	} else if strings.HasSuffix(book.Src, ".fb2") {
		if fb2Stream, err = os.Open(book.Src); err != nil {
			return
		}
		defer fb2Stream.Close()
	}

	decoder := xml.NewDecoder(fb2Stream)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&res)

	return
}
