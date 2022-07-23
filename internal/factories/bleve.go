package factories

import (
	"os"
	"path"

	"github.com/blevesearch/bleve/v2"
	blevemapping "github.com/blevesearch/bleve/v2/mapping"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

func NewBleveIndex(dir, name string, mapping blevemapping.IndexMapping) bleve.Index {
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	dir = path.Join(dir, name)

	db, err := bleve.Open(dir)
	if err != nil {
		db, err = bleve.New(dir, mapping)
		if err != nil {
			panic(err)
		}
	}

	return db
}
