package indexing

import (
	"fmt"
	"os"
	"path/filepath"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

func NewTmpIndex(
	src os.FileInfo, rootDir string, canReplace bool, fieldsMapping mapping.IndexMapping,
) (index entities.ISearchIndex, err error) {
	indexPath := filepath.Join(rootDir, src.Name(), fmt.Sprint(src.Size())) // @TODO: hashsum

	if _, err = os.Stat(indexPath); err == nil && !canReplace {
		return nil, fmt.Errorf("index for %s already exists", src.Name())
	}

	if err = os.RemoveAll(filepath.Dir(indexPath)); err != nil && os.IsNotExist(err) {
		return
	}

	return bleve.New(indexPath+tmpSuffix, fieldsMapping)
}
