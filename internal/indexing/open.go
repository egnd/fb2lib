package indexing

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	bleve "github.com/blevesearch/bleve/v2"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

func OpenIndex(dir string) (entities.ISearchIndex, error) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	items, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var indexes []bleve.Index
	for _, item := range items {
		if !item.IsDir() {
			continue
		}

		subitems, err := os.ReadDir(path.Join(dir, item.Name()))
		if err != nil {
			return nil, err
		}

		if len(subitems) == 0 || strings.HasSuffix(subitems[0].Name(), tmpSuffix) {
			continue
		}

		index, err := bleve.Open(filepath.Join(dir, item.Name(), subitems[0].Name()))
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, index)
	}

	return bleve.NewIndexAlias(indexes...), nil
}
