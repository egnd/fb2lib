package factories

import (
	bleve "github.com/blevesearch/bleve/v2"
	"github.com/spf13/viper"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

func NewBooksIndex(cfg *viper.Viper) (index bleve.Index, err error) {
	indexPath := cfg.GetString("bleve.path")

	if index, err = bleve.Open(indexPath); err == bleve.ErrorIndexPathDoesNotExist {
		mapping := bleve.NewIndexMapping() // @TODO:
		index, err = bleve.New(indexPath, mapping)
	}

	return
}
