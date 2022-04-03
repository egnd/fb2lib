package factories

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

func NewIndexMappingBook() *mapping.IndexMappingImpl {
	return bleve.NewIndexMapping() // @TODO:
}

func NewBooksIndex(name string, dir string) (bleve.Index, error) {
	indexPath := filepath.Join(dir,
		fmt.Sprintf("%x", md5.Sum([]byte(name))),
	)

	if err := os.RemoveAll(indexPath); err != nil && os.IsNotExist(err) {
		return nil, err
	}

	return bleve.New(indexPath, NewIndexMappingBook())
}

func OpenBooksIndex(dir string) (index bleve.Index, err error) {
	var indexes []bleve.Index

	items, err := os.ReadDir(dir)
	for _, item := range items {
		if !item.IsDir() {
			continue
		}

		if index, err = bleve.Open(
			filepath.Join(dir, item.Name()),
		); err != nil {
			return nil, err
		}

		indexes = append(indexes, index)
	}

	index = bleve.NewIndexAlias(indexes...)

	return
}
