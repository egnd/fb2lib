package entities

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type LibEncodeType string

const (
	LibEncodeMarshaler LibEncodeType = "marshaler"
	LibEncodeParser    LibEncodeType = "parser"
)

type Libraries map[string]Library

func (l *Libraries) GetSize() (res int64) {
	for _, item := range *l {
		if item.Disabled {
			continue
		}

		res += item.GetSize()
	}

	return
}

func (l *Libraries) GetItems() (res []LibItem, err error) {
	var filtered int

	for order := 0; filtered != len(*l); order++ {
		for _, lib := range *l {
			if lib.Order != order {
				continue
			}

			filtered++

			if lib.Disabled {
				continue
			}

			items, err := lib.GetItems()
			if err != nil {
				return nil, err
			}

			for _, libItem := range items {
				res = append(res, LibItem{Item: libItem, Lib: lib.Name})
			}
		}
	}

	return res, nil
}

type Library struct {
	Disabled bool `mapstructure:"disabled"`
	Order    int  `mapstructure:"order"`
	Name     string
	Dir      string        `mapstructure:"dir"`
	Index    string        `mapstructure:"index"`
	Types    []string      `mapstructure:"types"`
	Encoder  LibEncodeType `mapstructure:"encoder"`
}

func NewLibraries(cfgKey string, cfg *viper.Viper) (Libraries, error) {
	libs := Libraries{}

	if err := cfg.UnmarshalKey(cfgKey, &libs); err != nil {
		return nil, err
	}

	for name, lib := range libs {
		lib.Name = name
		if lib.Index == "" {
			lib.Index = lib.Name
		}

		libs[name] = lib
	}

	return libs, nil
}

func (l *Library) GetItems() (res []string, err error) {
	err = filepath.Walk(l.Dir, func(pathStr string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walk %s/%s error", l.Dir, pathStr)
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

type LibItem struct {
	Item string
	Lib  string
}
