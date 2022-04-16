package factories

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

// https://medevel.com/os-fulltext-search-solutions/
// https://habr.com/ru/post/333714/
// https://blevesearch.com

const (
	indexTmpSuffix = "_tmp"
)

func NewTmpIndex(
	src os.FileInfo, rootDir string, canReplace bool, fieldsMapping mapping.IndexMapping,
) (index entities.ISearchIndex, err error) {
	indexPath := filepath.Join(rootDir, src.Name(), fmt.Sprint(src.Size()))

	if _, err = os.Stat(indexPath); err == nil && !canReplace {
		return nil, fmt.Errorf("index for %s already exists", src.Name())
	}

	if err = os.RemoveAll(filepath.Dir(indexPath)); err != nil && os.IsNotExist(err) {
		return
	}

	return bleve.New(indexPath+indexTmpSuffix, fieldsMapping)
}

func NewIndex(
	name string, rootDir string, canReplace bool, fieldsMapping mapping.IndexMapping,
) (index entities.ISearchIndex, err error) {
	indexPath := filepath.Join(rootDir, name)

	if _, err = os.Stat(indexPath); err == nil {
		if canReplace {
			if err = os.RemoveAll(filepath.Dir(indexPath)); err != nil && os.IsNotExist(err) {
				return
			}
		} else {
			return bleve.Open(indexPath)
		}
	}

	return bleve.New(indexPath, fieldsMapping)
}

func OpenIndex(dir string, logger zerolog.Logger) (entities.ISearchIndex, error) {
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

		if len(subitems) == 0 || strings.HasSuffix(subitems[0].Name(), indexTmpSuffix) {
			continue
		}

		index, err := bleve.Open(filepath.Join(dir, item.Name(), subitems[0].Name()))
		if err != nil {
			return nil, err
		}

		logger.Debug().Str("dir", dir).Str("name", item.Name()).Msg("index opened")

		indexes = append(indexes, index)
	}

	return bleve.NewIndexAlias(indexes...), nil
}

func SaveTmpIndex(index entities.ISearchIndex) (err error) {
	if !strings.HasSuffix(index.Name(), indexTmpSuffix) {
		return
	}

	return os.Rename(index.Name(), strings.TrimSuffix(index.Name(), indexTmpSuffix))
}
