package library

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type LocalFSItems struct {
	rootDir    string
	extensions []string
	logger     zerolog.Logger
}

func NewLocalFSItems(path string, extensions []string, logger zerolog.Logger) *LocalFSItems {
	return &LocalFSItems{
		rootDir:    path,
		extensions: extensions,
		logger:     logger,
	}
}

func (l *LocalFSItems) allowedExt(ext string) bool {
	for _, v := range l.extensions {
		if v == ext {
			return true
		}
	}

	return false
}

func (l *LocalFSItems) GetAll() (items []string, err error) {
	err = filepath.Walk(l.rootDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if l.allowedExt(filepath.Ext(info.Name())) {
				items = append(items, path)
			}

			return nil
		})

	return
}

func (l *LocalFSItems) IterateItems(handler ILibItemHandler) (err error) {
	var files []string
	if files, err = l.GetAll(); err != nil {
		return
	}

	total := len(files)

	for fileNum, file := range files {
		logger := l.logger.With().Str("lib_item", file).Logger()

		finfo, err := os.Stat(file)
		if err != nil {
			logger.Error().Err(err).Msg("lib item stat")
		}

		if err := handler(file, finfo, fileNum+1, total, logger); err != nil {
			logger.Error().Err(err).Msg("lib item handler")
		}
	}

	return
}
