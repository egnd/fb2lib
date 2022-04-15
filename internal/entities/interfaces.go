package entities

import (
	"io"

	"github.com/blevesearch/bleve/v2"
	"gitlab.com/egnd/bookshelf/pkg/pagination"
)

type IIndexFactory func(string) (bleve.Index, error)

type IBooksRepo interface {
	SearchAll(string, pagination.IPager) ([]BookIndex, error)
	SearchByAuthor(string, pagination.IPager) ([]BookIndex, error)
	SearchBySequence(string, pagination.IPager) ([]BookIndex, error)
	GetBook(string) (BookIndex, error)
}

type ISearchIndex interface {
	io.Closer
	DocCount() (uint64, error)
	Search(req *bleve.SearchRequest) (*bleve.SearchResult, error)
	Index(id string, data interface{}) error
	Name() string
}
