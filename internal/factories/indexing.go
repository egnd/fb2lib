package factories

import (
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/spf13/viper"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

func NewBooksIndex(cfg *viper.Viper) (res entities.ISearchIndex, err error) {
	var libs entities.CfgLibsMap
	if libs, err = entities.NewCfgLibsMap(cfg, ""); err != nil {
		return
	}

	indexes := make([]bleve.Index, 0, len(libs))
	knownIndex := map[string]struct{}{}

	for _, lib := range libs {
		if !filepath.IsAbs(lib.IndexDir) {
			if lib.IndexDir, err = filepath.Abs(lib.IndexDir); err != nil {
				return
			}
		}

		if _, ok := knownIndex[lib.IndexDir]; ok {
			continue
		}

		knownIndex[lib.IndexDir] = struct{}{}

		var index bleve.Index
		if index, err = bleve.Open(lib.IndexDir); err != nil {
			return
		}

		indexes = append(indexes, index)
	}

	return bleve.NewIndexAlias(indexes...), nil
}
