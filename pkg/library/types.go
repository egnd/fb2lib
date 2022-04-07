package library

import (
	"archive/zip"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog"
)

type ILibItemHandler func(file os.FileInfo, dir string, num, total int, logger zerolog.Logger) error

type IZipItemHandler func(item *zip.File, data io.Reader, offset, num int64, logger zerolog.Logger) error

type ErrIterStop struct{}

func (e *ErrIterStop) Error() string {
	return "stop iteration"
}

type IHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type IExtractorFactory func(filePath string) (IZipExtractor, error)

type IZipExtractor interface {
	io.Closer
	GetSection(from, to int64) (io.ReadCloser, error)
}
