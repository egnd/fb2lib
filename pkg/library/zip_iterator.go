package library

import (
	"archive/zip"
	"compress/flate"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type ZipItemReader struct {
	path   string
	logger zerolog.Logger
}

func NewZipItemIterator(path string, logger zerolog.Logger) *ZipItemReader {
	return &ZipItemReader{
		path:   path,
		logger: logger,
	}
}

func (z *ZipItemReader) IterateItems(handler IZipItemHandler) (err error) {
	var archiveFile *os.File
	if archiveFile, err = os.Open(z.path); err != nil {
		return
	}
	defer archiveFile.Close()

	var archiveReader *zip.ReadCloser
	if archiveReader, err = zip.OpenReader(z.path); err != nil {
		return
	}
	defer archiveReader.Close()

	var i int64
	for _, zipFile := range archiveReader.File {
		i++
		logger := z.logger.With().Str("zip_item", zipFile.Name).Logger()

		if zipFile.Method != zip.Deflate {
			logger.Warn().Uint16("compression", zipFile.Method).Msg("check item compression type")
			continue
		}

		offset, err := zipFile.DataOffset()
		if err != nil {
			logger.Error().Err(err).Msg("readig zip item offset")
			continue
		}

		flateReader := flate.NewReader(io.NewSectionReader(archiveFile, offset, int64(zipFile.CompressedSize64)))
		if err = handler(zipFile, flateReader, offset, i, logger); err != nil {
			logger.Error().Err(err).Msg("handling zip item")
		}

		flateReader.Close()
	}

	return
}
