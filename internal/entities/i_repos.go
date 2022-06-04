package entities

import (
	"io"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/egnd/go-fb2parse"
)

type IBooksInfoRepo interface {
	io.Closer
	GetItems(query.Query, pagination.IPager, []search.SearchSort, *bleve.HighlightRequest, ...string) ([]BookInfo, error)
	FindByID(string) (BookInfo, error)
	FindIn(libName, queryStr string, pager pagination.IPager) ([]BookInfo, error)
	SaveBook(BookInfo) error
	GetGenresFreq(limit int) (GenresIndex, error)
	// SearchAll(string, pagination.IPager) ([]BookInfo, error)
	// SearchByAuthor(string, pagination.IPager) ([]BookInfo, error)
	// SearchBySequence(string, pagination.IPager) ([]BookInfo, error)
	// GetBook(string) (BookInfo, error)
	// GetGenres(int) (GenresIndex, error)
}

type IBooksLibraryRepo interface {
	GetFB2(BookInfo) (fb2parse.FB2File, error)
}

type ILibMarksRepo interface {
	MarkExists(string) bool
	AddMark(string) error
}
