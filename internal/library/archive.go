package library

import (
	"archive/zip"
	"compress/flate"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type ArchiveZip struct {
	path   string
	logger zerolog.Logger
}

func NewArchiveZip(path string) *ArchiveZip {
	return &ArchiveZip{
		path: path,
	}
}

func (a *ArchiveZip) Walk(logger zerolog.Logger, handler ArchiveItemHandler) (err error) {
	var archiveFile *os.File
	if archiveFile, err = os.Open(a.path); err != nil {
		return
	}
	defer archiveFile.Close()

	var archiveReader *zip.ReadCloser
	if archiveReader, err = zip.OpenReader(a.path); err != nil {
		return
	}
	defer archiveReader.Close()

	var i int64
	for _, zipFile := range archiveReader.File {
		i++

		if zipFile.Method != zip.Deflate {
			logger.Warn().
				Uint16("method", zipFile.Method).
				Str("file", zipFile.Name).
				Msg("not deflate compression")

			continue
		}

		offset, _ := zipFile.DataOffset()

		err = func() error {
			flateReader := flate.NewReader(
				io.NewSectionReader(archiveFile, offset, int64(zipFile.CompressedSize64)),
			)
			defer flateReader.Close()

			return handler(zipFile, flateReader, offset, i)
		}()

		if err != nil {
			return err
		}
	}

	return
}

type ArchiveItemHandler func(file *zip.File, data io.Reader, offset, num int64) error
