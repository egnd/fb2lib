package tasks

import (
	"fmt"
	"io"
	"sync"

	"github.com/rs/zerolog"

	"github.com/egnd/fb2lib/internal/entities"
)

type IndexFB2DataTask struct {
	data    io.Reader
	book    entities.BookInfo
	encoder entities.LibEncodeType
	repo    entities.IBooksInfoRepo
	logger  zerolog.Logger
	wg      *sync.WaitGroup
	counter *entities.CntAtomic32
}

func NewIndexFB2DataTask(
	data io.Reader,
	book entities.BookInfo,
	encoder entities.LibEncodeType,
	repo entities.IBooksInfoRepo,
	logger zerolog.Logger,
	wg *sync.WaitGroup,
	counter *entities.CntAtomic32,
) *IndexFB2DataTask {
	return &IndexFB2DataTask{
		data:    data,
		book:    book,
		encoder: encoder,
		repo:    repo,
		wg:      wg,
		counter: counter,
		logger:  logger.With().Str("task", "index_fb2_data").Logger(),
	}
}

func (t *IndexFB2DataTask) GetID() string {
	return fmt.Sprintf("index_fb2_data [%s] %s", t.book.LibName, t.book.Src)
}

func (t *IndexFB2DataTask) Do() {
	defer t.wg.Done()
	// defer func() {
	// 	if t.bar != nil {
	// 		t.bar.IncrInt64(int64(t.file.CompressedSize64))
	// 	}
	// }()

	fb2File, err := entities.ParseFB2(t.data, t.encoder,
		SkipFB2Binaries, SkipFB2DocInfo, SkipFB2CustomInfo, SkipFB2Cover,
	)
	if err != nil {
		t.logger.Error().Err(err).Msg("parse fb2 data")
		return
	}

	t.book.Index = entities.NewFB2Index(&fb2File)

	if err = t.repo.SaveBook(t.book); err != nil {
		t.logger.Error().Err(err).Msg("index fb2 data")
		return
	}

	t.counter.Inc(1)
	t.logger.Debug().Msg("index")
}
