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

	index, err := bleve.Open(dir)
	if err != nil {
		index, err = bleve.New(dir, mapping)
		if err != nil {
			panic(err)
		}
	}

	return index
}

// func NewCompositeBleveIndex(dir string,
// 	libs entities.Libraries, mapping blevemapping.IndexMapping,
// ) bleve.Index {
// 	indexes := make([]bleve.Index, 0, len(libs))
// 	opened := map[string]struct{}{}

// 	for _, lib := range libs {
// 		if _, ok := opened[lib.Index]; ok || lib.Disabled {
// 			continue
// 		}

// 		opened[lib.Index] = struct{}{}
// 		indexes = append(indexes, NewBleveIndex(dir, lib.Index, mapping))
// 	}

// 	return bleve.NewIndexAlias(indexes...)
// }
