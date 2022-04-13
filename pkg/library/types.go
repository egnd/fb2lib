package library

import (
	"archive/zip"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type ILibItemHandler func(file os.FileInfo, dir string, num, total int, logger zerolog.Logger) error

type IZipItemHandler func(item *zip.File, data io.Reader, offset, num int64, logger zerolog.Logger) error

type ErrIterStop struct{}

func (e *ErrIterStop) Error() string {
	return "stop iteration"
}
