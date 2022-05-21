package entities

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Libraries map[string]Library

func (l *Libraries) GetSize() (res int64) {
	for _, item := range *l {
		res += item.GetSize()
	}

	return
}

func (l *Libraries) GetItems() (map[string]string, error) {
	res := map[string]string{}

	for _, lib := range *l {
		items, err := lib.GetItems()
		if err != nil {
			return nil, err
		}

		for _, libItem := range items {
			res[libItem] = lib.Name
		}
	}

	return res, nil
}

type Library struct {
	Disabled bool   `mapstructure:"disabled"`
	BooksDir string `mapstructure:"books_dir"`
	IndexDir string `mapstructure:"index_dir"`
	Name     string
	Types    []string `mapstructure:"types"`
}

func NewLibraries(cfgKey string, cfg *viper.Viper) (Libraries, error) {
	libs := Libraries{}

	if err := cfg.UnmarshalKey(cfgKey, &libs); err != nil {
		return nil, err
	}

	for name, lib := range libs {
		if lib.Disabled {
			delete(libs, name)
			continue
		}

		lib.Name = name
		libs[name] = lib
	}

	return libs, nil
}

func (l *Library) GetItems() (res []string, err error) {
	err = filepath.Walk(l.BooksDir,
		func(pathStr string, info os.FileInfo, iterErr error) error {
			if iterErr != nil {
				return iterErr
			}

			if info.IsDir() || !SliceHasString(l.Types, strings.TrimPrefix(path.Ext(info.Name()), ".")) {
				return nil
			}

			res = append(res, pathStr)

			return nil
		})

	return
}

func (l *Library) GetSize() int64 {
	items, err := l.GetItems()
	if err != nil {
		panic(err)
	}

	var res int64
	for _, itemPath := range items {
		info, err := os.Stat(itemPath)
		if err != nil {
			panic(err)
		}

		res += info.Size()
	}

	return res
}
