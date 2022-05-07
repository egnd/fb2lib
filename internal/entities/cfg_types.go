package entities

import "github.com/spf13/viper"

type CfgLibsMap map[string]CfgLibrary

func NewCfgLibsMap(cfg *viper.Viper, specificLib string) (libs CfgLibsMap, err error) {
	if err = cfg.UnmarshalKey("libraries", &libs); err != nil {
		return
	}

	if lib, ok := libs[specificLib]; specificLib != "" && ok {
		return CfgLibsMap{specificLib: lib}, nil
	}

	for name, lib := range libs {
		if !lib.Enabled {
			delete(libs, name)
		}
	}

	return
}

type CfgLibrary struct {
	Enabled  bool   `mapstructure:"enabled"`
	BooksDir string `mapstructure:"books_dir"`
	IndexDir string `mapstructure:"index_dir"`
}
