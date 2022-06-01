package tasks

import (
	"os"
	"path"
	"strings"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-wpool/v2/interfaces"
	"github.com/rs/zerolog"
)

type HandleLibItemTask struct {
	item      string
	lib       entities.Library
	logger    zerolog.Logger
	readPool  interfaces.Pool
	indexPool interfaces.Pool
	wg        *sync.WaitGroup
	taskIndex IndexTaskFactory
	repoMarks entities.ILibMarksRepo
	counter   *entities.CntAtomic32
}

func NewHandleLibItemTask(
	item string,
	lib entities.Library,
	logger zerolog.Logger,
	readPool interfaces.Pool,
	indexPool interfaces.Pool,
	wg *sync.WaitGroup,
	repoMarks entities.ILibMarksRepo,
	counter *entities.CntAtomic32,
	taskIndex IndexTaskFactory,
) *HandleLibItemTask {
	return &HandleLibItemTask{
		wg:        wg,
		lib:       lib,
		item:      item,
		readPool:  readPool,
		indexPool: indexPool,
		taskIndex: taskIndex,
		repoMarks: repoMarks,
		counter:   counter,
		logger: logger.With().Str("libname", lib.Name).
			Str("libitem", strings.TrimPrefix(item, lib.Dir)).Logger(),
	}
}

func (t *HandleLibItemTask) GetID() string {
	return "handle_lib_item"
}

func (t *HandleLibItemTask) Do() {
	finfo, err := os.Stat(t.item)
	if err != nil {
		t.logger.Error().Err(err).Msg("stat lib item")
		return
	}

	if t.repoMarks.MarkExists(t.item) {
		t.logger.Info().Msg("already indexed")

		// if t.bar != nil {
		// 	t.bar.IncrInt64(finfo.Size())
		// }

		return
	}

	switch path.Ext(t.item) {
	case ".zip":
		NewReadZipTask(t.item, finfo, t.lib, t.repoMarks, t.logger, t.readPool, t.indexPool, t.wg, t.counter, t.taskIndex).Do()

		if err := t.repoMarks.AddMark(t.item); err != nil {
			t.logger.Error().Err(err).Msg("memorize zip file")
		}
	case ".fb2":
		t.wg.Add(1)
		t.counter.Inc(1)

		reader, err := os.Open(t.item)
		if err != nil {
			t.logger.Error().Err(err).Msg("open fb2 file")
			return
		}

		if err := t.readPool.AddTask(
			NewReaderTask(reader, entities.BookInfo{
				LibName: t.lib.Name,
				Size:    uint64(finfo.Size()),
				Src:     strings.TrimPrefix(t.item, t.lib.Dir),
			}, t.wg, t.logger, t.indexPool, t.taskIndex),
		); err != nil {
			t.logger.Error().Err(err).Msg("send fb2 to pool")
			return
		}

		t.logger.Debug().Msg("iterate")
	default:
		t.logger.Warn().Str("type", path.Ext(t.item)).Msg("invalid lib item type")
	}
}
