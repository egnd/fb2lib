package repos

import (
	"compress/flate"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"golang.org/x/net/html/charset"
)

type BooksDataFB2Files struct {
}

func NewBooksDataFB2Files() *BooksDataFB2Files {
	return &BooksDataFB2Files{}
}

func (r *BooksDataFB2Files) GetFor(book entities.BookIndex) (res entities.FB2Book, err error) {
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
