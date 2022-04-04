package library

import (
	"archive/zip"
	"compress/flate"
	"io"
	"os"
)

type ZipExtractorLocal struct {
	file   *os.File
	reader *zip.ReadCloser
}

func FactoryZipExtractorLocal() IExtractorFactory {
	return func(filePath string) (IZipExtractor, error) {
		return NewZipExtractorLocal(filePath)
	}
}

func NewZipExtractorLocal(path string) (res *ZipExtractorLocal, err error) {
	res = &ZipExtractorLocal{}
	res.file, err = os.Open(path)

	return
}

func (z *ZipExtractorLocal) Close() error {
	return z.file.Close()
}

func (z *ZipExtractorLocal) GetSection(from int64, to int64) (res io.ReadCloser, err error) {
	return flate.NewReader(io.NewSectionReader(z.file, from, to)), nil
}
