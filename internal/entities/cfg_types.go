package entities

import (
	"github.com/spf13/viper"
)

type CfgLibsMap map[string]CfgLibrary

func NewCfgLibsMap(cfg *viper.Viper, specificLib string) (libs CfgLibsMap, err error) {
	if err = cfg.UnmarshalKey("libraries", &libs); err != nil {
		return
	}

	if lib, ok := libs[specificLib]; specificLib != "" && ok {
		lib.Name = specificLib

		return CfgLibsMap{specificLib: lib}, nil
	}

	for name, lib := range libs {
		if !lib.Enabled {
			delete(libs, name)
			continue
		}

		lib.Name = name
		libs[name] = lib
	}

	return
}

type CfgLibrary struct {
	Enabled  bool `mapstructure:"enabled"`
	Name     string
	BooksDir string `mapstructure:"books_dir"`
	IndexDir string `mapstructure:"index_dir"`
}
