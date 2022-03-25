package factories

import (
	"path/filepath"

	bleve "github.com/blevesearch/bleve/v2"
	"github.com/spf13/viper"
)

func NewBleveDB(cfg *viper.Viper) (index bleve.Index, err error) {
	return bleve.New(
		cfg.GetString(filepath.Join(cfg.GetString("storage.dir"), cfg.GetString("storage.index_file"))),
		bleve.NewIndexMapping(),
	)
}
