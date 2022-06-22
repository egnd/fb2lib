package entities

import (
	"io"

	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/egnd/go-xmlparse/fb2"
)

type IBooksInfoRepo interface {
	io.Closer
	GetByID(bookID string) (*Book, error)
	FindBooks(queryStr string, idxField IndexField, idxFieldVal string, pager pagination.IPager) ([]Book, error)
	Remove(bookID string) error
	GetSeriesBooks(limit int, series []string, except *Book) (res []Book, err error)
	GetAuthorsBooks(limit int, authors []string, except *Book) (res []Book, err error)
	GetAuthorsSeries(authors []string, except []string) (res map[string]int, err error)
	GetBooksCnt() (uint64, error)
	GetAuthorsCnt() (uint64, error)
	GetGenresCnt() (uint64, error)
	GetSeriesCnt() (uint64, error)
	GetGenres(pager pagination.IPager) (FreqsItems, error)
	GetSeriesByChar(char rune) (FreqsItems, error)
	GetAuthorsByChar(char rune) (FreqsItems, error)
	SaveBook(book *Book) error
}

type IBooksLibraryRepo interface {
	GetFB2(Book) (fb2.File, error)
}

type ILibMarksRepo interface {
	MarkExists(string) bool
	AddMark(string) error
}
