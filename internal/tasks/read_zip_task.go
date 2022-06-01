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
	"github.com/egnd/go-wpool/v2/interfaces"
	"github.com/rs/zerolog"
)

type ReadZipTask struct {
	path      string
	item      fs.FileInfo
	lib       entities.Library
	repoMarks entities.ILibMarksRepo
	logger    zerolog.Logger
	readPool  interfaces.Pool
	indexPool interfaces.Pool
	wg        *sync.WaitGroup
	taskIndex IndexTaskFactory
	counter   *entities.CntAtomic32
}

func NewReadZipTask(
	path string,
	item fs.FileInfo,
	lib entities.Library,
	repoMarks entities.ILibMarksRepo,
	logger zerolog.Logger,
	readPool interfaces.Pool,
	indexPool interfaces.Pool,
	wg *sync.WaitGroup,
	counter *entities.CntAtomic32,
	taskIndex IndexTaskFactory,
) *ReadZipTask {
	return &ReadZipTask{
		path:      path,
		item:      item,
		lib:       lib,
		repoMarks: repoMarks,
		readPool:  readPool,
		indexPool: indexPool,
		wg:        wg,
		taskIndex: taskIndex,
		counter:   counter,
		logger:    logger.With().Str("task", "read_zip").Logger(),
	}
}

func (t *ReadZipTask) GetID() string {
	return fmt.Sprintf("read_zip %s", t.item.Name())
}

func (t *ReadZipTask) Do() {
	archive, err := os.Open(t.path)
	if err != nil {
		t.logger.Error().Err(err).Msg("open zip file")
		return
	}
	// defer archive.Close()

	itemReader, err := zip.OpenReader(t.path)
	if err != nil {
		t.logger.Error().Err(err).Msg("read zip file")
		return
	}
	defer itemReader.Close()

	for _, book := range itemReader.File {
		logger := t.logger.With().Str("libsubitem", book.Name).Logger()

		offset, err := book.DataOffset()
		if err != nil {
			logger.Error().Err(err).Msg("get offset")
			return
		}

		var reader io.ReadCloser
		switch book.Method {
		case zip.Deflate:
			reader = flate.NewReader(io.NewSectionReader(archive, offset, int64(book.CompressedSize64)))
		case zip.Store:
			reader = io.NopCloser(io.NewSectionReader(archive, offset, int64(book.CompressedSize64)))
		default:
			logger.Warn().Uint16("method", book.Method).Msg("undefined compression type")
			return
		}

		t.wg.Add(1)
		t.counter.Inc(1)

		if err := t.readPool.AddTask(
			NewReaderTask(reader, entities.BookInfo{
				Offset:         uint64(offset),
				Size:           book.UncompressedSize64,
				SizeCompressed: book.CompressedSize64,
				LibName:        t.lib.Name,
				Src:            path.Join(strings.TrimPrefix(t.path, t.lib.Dir), book.Name),
			}, t.wg, logger, t.indexPool, t.taskIndex),
		); err != nil {
			logger.Error().Err(err).Msg("send fb2 to pool")
			return
		}

		logger.Debug().Msg("iterate")
	}
}
