package factories

import (
	"path"

	"github.com/blevesearch/bleve/v2"
	blevemapping "github.com/blevesearch/bleve/v2/mapping"
	"github.com/egnd/fb2lib/internal/entities"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

func NewBleveIndex(
	dir, libName string, mapping blevemapping.IndexMapping,
) bleve.Index {
	index, err := bleve.Open(path.Join(dir, libName))
	if err != nil {
		index, err = bleve.New(path.Join(dir, libName), mapping)
		if err != nil {
			panic(err)
		}
	}

	return index
}

func NewCompositeBleveIndex(dir string,
	libs entities.Libraries, mapping blevemapping.IndexMapping,
) bleve.Index {
	indexes := make([]bleve.Index, 0, len(libs))

	for libName := range libs {
		indexes = append(indexes, NewBleveIndex(dir, libName, mapping))
	}

	return bleve.NewIndexAlias(indexes...)
}
