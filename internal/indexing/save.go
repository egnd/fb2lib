package indexing

import (
	"os"
	"strings"

	"gitlab.com/egnd/bookshelf/internal/entities"
)

func SaveTmpIndex(index entities.ISearchIndex) (err error) {
	if !strings.HasSuffix(index.Name(), tmpSuffix) {
		return
	}

	return os.Rename(index.Name(), strings.TrimSuffix(index.Name(), tmpSuffix))
}
