package entities

import (
	"io"

	"github.com/blevesearch/bleve/v2"
)

type ISearchIndex interface {
	io.Closer
	DocCount() (uint64, error)
	Search(req *bleve.SearchRequest) (*bleve.SearchResult, error)
	Index(id string, data interface{}) error
	Name() string
	NewBatch() *bleve.Batch
	Batch(b *bleve.Batch) error
}

type IMarshal func(any) ([]byte, error)

type IUnmarshal func([]byte, any) error
