package tasks

import (
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"

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
	bar     *mpb.Bar
}

func NewIndexFB2DataTask(
	data io.Reader,
	book entities.BookInfo,
	encoder entities.LibEncodeType,
	repo entities.IBooksInfoRepo,
	logger zerolog.Logger,
	wg *sync.WaitGroup,
	counter *entities.CntAtomic32,
	bar *mpb.Bar,
) *IndexFB2DataTask {
	return &IndexFB2DataTask{
		data:    data,
		book:    book,
		encoder: encoder,
		repo:    repo,
		wg:      wg,
		counter: counter,
		bar:     bar,
		logger:  logger.With().Str("task", "index_fb2_data").Logger(),
	}
}

func (t *IndexFB2DataTask) ID() string {
	return fmt.Sprintf("index_fb2_data [%s] %s", t.book.LibName, t.book.Src)
}

func (t *IndexFB2DataTask) Do() error {
	defer t.wg.Done()

	if t.bar != nil {
		if t.book.SizeCompressed > 0 {
			defer t.bar.IncrInt64(int64(t.book.SizeCompressed))
		} else {
			defer t.bar.IncrInt64(int64(t.book.Size))
		}
	}

	fb2File, err := entities.ParseFB2(t.data, t.encoder,
		SkipFB2Binaries, SkipFB2DocInfo, SkipFB2CustomInfo, SkipFB2Cover,
	)
	if err != nil {
		return errors.Wrap(err, "parse fb2 data")
	}

	t.book.Index = entities.NewFB2Index(&fb2File)
	t.book.Index.Lib = t.book.LibName

	if err = t.repo.SaveBook(t.book); err != nil {
		return errors.Wrap(err, "index fb2 data")
	}

	t.counter.Inc(1)
	t.logger.Debug().Msg("index")

	return nil
}
