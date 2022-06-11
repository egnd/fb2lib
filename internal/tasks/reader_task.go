package tasks

import (
	"bytes"
	"io"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-pipeline"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type ReaderTask struct {
	reader    io.ReadCloser
	book      entities.BookInfo
	wg        *sync.WaitGroup
	logger    zerolog.Logger
	indexPool pipeline.Dispatcher
	taskIndex IndexTaskFactory
}

func NewReaderTask(
	reader io.ReadCloser,
	book entities.BookInfo,
	wg *sync.WaitGroup,
	logger zerolog.Logger,
	indexPool pipeline.Dispatcher,
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

func (t *ReaderTask) ID() string {
	return "reader"
}

func (t *ReaderTask) Do() error {
	defer t.wg.Done()
	defer t.reader.Close()

	data, err := io.ReadAll(t.reader)
	if err != nil {
		return errors.Wrap(err, "read lib item file")
	}

	t.wg.Add(1)

	if err = t.indexPool.Push(t.taskIndex(bytes.NewBuffer(data), t.book, t.logger)); err != nil {
		return errors.Wrap(err, "send data to index pool")
	}

	t.logger.Debug().Msg("read")

	return nil
}
