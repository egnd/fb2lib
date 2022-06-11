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
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-pipeline"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

type ReadZipTask struct {
	num       int
	total     int
	path      string
	item      fs.FileInfo
	lib       entities.Library
	repoMarks entities.ILibMarksRepo
	logger    zerolog.Logger
	readPool  pipeline.Dispatcher
	indexPool pipeline.Dispatcher
	wg        *sync.WaitGroup
	taskIndex IndexTaskFactory
	counter   *entities.CntAtomic32
	bars      *mpb.Progress
}

func NewReadZipTask(
	num, total int,
	path string,
	item fs.FileInfo,
	lib entities.Library,
	repoMarks entities.ILibMarksRepo,
	logger zerolog.Logger,
	readPool pipeline.Dispatcher,
	indexPool pipeline.Dispatcher,
	wg *sync.WaitGroup,
	counter *entities.CntAtomic32,
	bars *mpb.Progress,
	taskIndex IndexTaskFactory,
) *ReadZipTask {
	return &ReadZipTask{
		num:       num,
		total:     total,
		path:      path,
		item:      item,
		lib:       lib,
		repoMarks: repoMarks,
		readPool:  readPool,
		indexPool: indexPool,
		wg:        wg,
		taskIndex: taskIndex,
		counter:   counter,
		bars:      bars,
		logger:    logger.With().Str("task", "read_zip").Logger(),
	}
}

func (t *ReadZipTask) ID() string {
	return fmt.Sprintf("read_zip %s", t.item.Name())
}

func (t *ReadZipTask) Do() error {
	archive, err := os.Open(t.path)
	if err != nil {
		return errors.Wrap(err, "open zip file")
	}
	// defer archive.Close()

	itemReader, err := zip.OpenReader(t.path)
	if err != nil {
		return errors.Wrap(err, "read zip file")
	}
	defer itemReader.Close()

	var bar *mpb.Bar
	if t.bars != nil {
		bar = t.initBar()
		defer bar.Abort(true)
	}

	for _, book := range itemReader.File {
		if err := func() error {
			if bar != nil {
				defer bar.IncrInt64(int64(book.CompressedSize64))
			}

			logger := t.logger.With().Str("libsubitem", book.Name).Logger()
			t.counter.Inc(1)

			offset, err := book.DataOffset()
			if err != nil {
				return errors.Wrap(err, "get offset")
			}

			reader, err := t.initReader(book, archive, offset)
			if err != nil {
				return errors.Wrap(err, "open subitem")
			}

			t.wg.Add(1)

			if err := t.readPool.Push(t.readerTask(reader, book, offset, logger)); err != nil {
				return errors.Wrap(err, "send fb2 to pool")
			}

			logger.Debug().Msg("iterate")

			return nil
		}(); err != nil {
			return err
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
		err = fmt.Errorf("undefined compression type %d", file.Method)
	}

	return
}

func (t *ReadZipTask) readerTask(reader io.ReadCloser, file *zip.File, offset int64, logger zerolog.Logger) pipeline.Task {
	return NewReaderTask(reader, entities.BookInfo{
		Offset:         uint64(offset),
		Size:           file.UncompressedSize64,
		SizeCompressed: file.CompressedSize64,
		LibName:        t.lib.Name,
		Src:            path.Join(strings.TrimPrefix(t.path, t.lib.Dir), file.Name),
	}, t.wg, logger, t.indexPool, t.taskIndex)
}
