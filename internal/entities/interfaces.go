package entities

import (
	"context"
	"io"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/blevesearch/bleve/v2"
)

type IIndexFactory func(string) (bleve.Index, error)

type IBooksRepo interface {
	GetBooks(context.Context, string, *pagination.Paginator) ([]BookIndex, error)
	GetBook(context.Context, string) (BookIndex, error)
}

type ISearchIndex interface {
	io.Closer
	DocCount() (uint64, error)
	Search(req *bleve.SearchRequest) (*bleve.SearchResult, error)
	Index(id string, data interface{}) error
	Name() string
}
