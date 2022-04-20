package entities

import (
	"io"

	"github.com/blevesearch/bleve/v2"
)

type IIndexFactory func(string) (bleve.Index, error)

type ISearchIndex interface {
	io.Closer
	DocCount() (uint64, error)
	Search(req *bleve.SearchRequest) (*bleve.SearchResult, error)
	Index(id string, data interface{}) error
	Name() string
	NewBatch() *bleve.Batch
	Batch(b *bleve.Batch) error
}
