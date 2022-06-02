package tasks

import (
	"bytes"
	"io"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-wpool/v2/interfaces"
	"github.com/rs/zerolog"
)

type ReaderTask struct {
	reader    io.ReadCloser
	book      entities.BookInfo
	wg        *sync.WaitGroup
	logger    zerolog.Logger
	indexPool interfaces.Pool
	taskIndex IndexTaskFactory
}

func NewReaderTask(
	reader io.ReadCloser,
	book entities.BookInfo,
	wg *sync.WaitGroup,
	logger zerolog.Logger,
	indexPool interfaces.Pool,
	taskIndex IndexTaskFactory,
) *ReaderTask {
	return &ReaderTask{
		book:      book,
		reader:    reader,
		wg:        wg,
		indexPool: indexPool,
		taskIndex: taskIndex,
		logger:    logger.With().Str("task", "reader").Logger(),
	}
}

func (t *ReaderTask) GetID() string {
	return "reader"
}

func (t *ReaderTask) Do() {
	defer t.wg.Done()
	defer t.reader.Close()

	data, err := io.ReadAll(t.reader)
	if err != nil {
		t.logger.Error().Err(err).Msg("read lib item file")
		return
	}

	t.wg.Add(1)

	if err = t.indexPool.AddTask(t.taskIndex(bytes.NewBuffer(data), t.book, t.logger)); err != nil {
		t.logger.Error().Err(err).Msg("send data to index pool")
		return
	}

	t.logger.Debug().Msg("read")
}
