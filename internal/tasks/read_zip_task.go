package tasks

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

type DoReadZipTask func(fs.FileInfo) error

type ReadZipTask struct {
	num          int
	total        int
	id           string
	path         string
	item         fs.FileInfo
	lib          entities.Library
	bars         *mpb.Progress
	doReaderTask PushReadTask
}

func NewReadZipTask(
	num int,
	total int,
	path string,
	item fs.FileInfo,
	lib entities.Library,
	bars *mpb.Progress,
	doReaderTask PushReadTask,
) *ReadZipTask {
	return &ReadZipTask{
		id:           fmt.Sprintf("iterate zip [%s] %s", lib.Name, strings.TrimPrefix(path, lib.Dir)),
		num:          num,
		total:        total,
		path:         path,
		item:         item,
		lib:          lib,
		bars:         bars,
		doReaderTask: doReaderTask,
	}
}

func (t *ReadZipTask) ID() string {
	return t.id
}

func (t *ReadZipTask) Do() error {
	archive, err := os.Open(t.path)
	if err != nil {
		return errors.Wrap(err, "open zip error")
	}
	// defer archive.Close()

	itemReader, err := zip.OpenReader(t.path)
	if err != nil {
		return errors.Wrap(err, "read zip error")
	}
	defer itemReader.Close()

	var bar *mpb.Bar
	if t.bars != nil {
		bar = t.initBar()
		defer bar.Abort(true)
	}

	for _, book := range itemReader.File {
		offset, err := book.DataOffset()
		if err != nil {
			return errors.Wrap(err, "offset error")
		}

		reader, err := t.initReader(book, archive, offset)
		if err != nil {
			return errors.Wrap(err, "open item error")
		}

		if err := t.doReaderTask(reader, entities.Book{
			Offset:         uint64(offset),
			Size:           book.UncompressedSize64,
			SizeCompressed: book.CompressedSize64,
			Lib:            t.lib.Name,
			Src:            path.Join(strings.TrimPrefix(t.path, t.lib.Dir), book.Name),
		}); err != nil {
			return errors.Wrap(err, "do item error")
		}

		if bar != nil {
			bar.IncrInt64(int64(book.CompressedSize64))
		}
	}

	return nil
}

func (t *ReadZipTask) initBar() *mpb.Bar {
	return t.bars.AddBar(t.item.Size(),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(decor.Name(fmt.Sprintf("[%d of %d] %s:%s",
			t.num, t.total, t.lib.Name, strings.TrimPrefix(t.path, t.lib.Dir),
		))),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f/% .2f"), decor.Name(", "),
			decor.AverageSpeed(decor.UnitKB, "% .2f"), decor.Name(", "),
			decor.AverageETA(decor.ET_STYLE_GO),
		),
	)
}

func (t *ReadZipTask) initReader(file *zip.File, archive io.ReaderAt, from int64) (reader io.ReadCloser, err error) {
	switch file.Method {
	case zip.Deflate:
		reader = flate.NewReader(io.NewSectionReader(archive, from, int64(file.CompressedSize64)))
	case zip.Store:
		reader = io.NopCloser(io.NewSectionReader(archive, from, int64(file.CompressedSize64)))
	default:
		err = fmt.Errorf("undefined compression method %d", file.Method)
	}

	return
}
