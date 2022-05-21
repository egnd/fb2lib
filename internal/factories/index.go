package factories

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/egnd/fb2lib/internal/entities"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

func NewBleveIndex(
	pathStr string, mapping mapping.IndexMapping,
) (index bleve.Index, err error) {
	if index, err = bleve.Open(pathStr); err != nil {
		index, err = bleve.New(pathStr, mapping)
	}

	return
}

func NewCompositeBleveIndex(
	libs entities.Libraries, mapping mapping.IndexMapping,
) (bleve.Index, error) {
	indexes := make([]bleve.Index, 0, len(libs))
	knownIndex := map[string]struct{}{}

	for _, lib := range libs {
		if _, ok := knownIndex[lib.IndexDir]; ok {
			continue
		}

		knownIndex[lib.IndexDir] = struct{}{}

		index, err := NewBleveIndex(lib.IndexDir, mapping)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, index)
	}

	return bleve.NewIndexAlias(indexes...), nil
}
