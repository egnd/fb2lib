package library

import (
	"archive/zip"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type ILibItemHandler func(path string, info os.FileInfo, num, total int, logger zerolog.Logger) error

type IZipItemHandler func(item *zip.File, data io.Reader, offset, num int64, logger zerolog.Logger) error
