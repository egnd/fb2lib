package tasks

import (
	"os"
	"path"
	"strings"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-pipeline"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
)

type HandleLibItemTask struct {
	num       int
	total     int
	item      string
	lib       entities.Library
	logger    zerolog.Logger
	readPool  pipeline.Dispatcher
	indexPool pipeline.Dispatcher
	wg        *sync.WaitGroup
	taskIndex IndexTaskFactory
	repoMarks entities.ILibMarksRepo
	counter   *entities.CntAtomic32
	bars      *mpb.Progress
}

func NewHandleLibItemTask(
	num, total int,
	item string,
	lib entities.Library,
	logger zerolog.Logger,
	readPool pipeline.Dispatcher,
	indexPool pipeline.Dispatcher,
	wg *sync.WaitGroup,
	repoMarks entities.ILibMarksRepo,
	counter *entities.CntAtomic32,
	bars *mpb.Progress,
	taskIndex IndexTaskFactory,
) *HandleLibItemTask {
	return &HandleLibItemTask{
		num:       num,
		total:     total,
		wg:        wg,
		lib:       lib,
		item:      item,
		readPool:  readPool,
		indexPool: indexPool,
		taskIndex: taskIndex,
		repoMarks: repoMarks,
		counter:   counter,
		bars:      bars,
		logger: logger.With().Str("libname", lib.Name).
			Str("libitem", strings.TrimPrefix(item, lib.Dir)).Logger(),
	}
}

func (t *HandleLibItemTask) ID() string {
	return "handle_lib_item"
}

func (t *HandleLibItemTask) Do() error {
	finfo, err := os.Stat(t.item)
	if err != nil {
		return errors.Wrap(err, "stat lib item")
	}

	if t.repoMarks.MarkExists(t.item) {
		t.logger.Info().Msg("already indexed")

		// if t.bar != nil {
		// 	t.bar.IncrInt64(finfo.Size())
		// }

		return nil
	}

	switch path.Ext(t.item) {
	case ".zip":
		NewReadZipTask(t.num, t.total, t.item, finfo, t.lib,
			t.repoMarks, t.logger, t.readPool, t.indexPool, t.wg, t.counter, t.bars, t.taskIndex,
		).Do()

		if err := t.repoMarks.AddMark(t.item); err != nil {
			t.logger.Error().Err(err).Msg("memorize zip file")
		}
	case ".fb2":
		t.wg.Add(1)
		t.counter.Inc(1)

		reader, err := os.Open(t.item)
		if err != nil {
			return errors.Wrap(err, "open fb2 file")
		}

		if err := t.readPool.Push(
			NewReaderTask(reader, entities.BookInfo{
				LibName: t.lib.Name,
				Size:    uint64(finfo.Size()),
				Src:     strings.TrimPrefix(t.item, t.lib.Dir),
			}, t.wg, t.logger, t.indexPool, t.taskIndex),
		); err != nil {
			return errors.Wrap(err, "send fb2 to pool")
		}

		t.logger.Debug().Msg("iterate")
	default:
		t.logger.Warn().Str("type", path.Ext(t.item)).Msg("invalid lib item type")
	}

	return nil
}
