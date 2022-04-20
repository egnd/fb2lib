package tasks

import (
	"context"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/pkg/fb2parser"
)

type FB2IndexTask struct {
	fb2Dir     string
	fb2File    os.FileInfo
	logger     zerolog.Logger
	wg         *sync.WaitGroup
	cntTotal   *entities.CntAtomic32
	cntIndexed *entities.CntAtomic32
	index      entities.ISearchIndex
	totalBar   *mpb.Bar
}

func NewFB2IndexTask(
	fb2File os.FileInfo,
	fb2Dir string,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	logger zerolog.Logger,
	wg *sync.WaitGroup,
	totalBar *mpb.Bar,
	index entities.ISearchIndex,
) *FB2IndexTask {
	return &FB2IndexTask{
		fb2Dir:     fb2Dir,
		fb2File:    fb2File,
		logger:     logger,
		wg:         wg,
		cntTotal:   cntTotal,
		cntIndexed: cntIndexed,
		totalBar:   totalBar,
		index:      index,
	}
}

func (t *FB2IndexTask) Do(context.Context) {
	defer func() {
		if t.totalBar != nil {
			t.totalBar.IncrInt64(t.fb2File.Size())
		}

		t.cntTotal.Inc(1)
		t.wg.Done()
	}()

	startTS := time.Now()

	fileReader, err := os.Open(path.Join(t.fb2Dir, t.fb2File.Name()))
	if err != nil {
		t.logger.Error().Err(err).Msg("reading fb2 file")
	}

	if t.indexFB2File(fileReader) {
		t.cntIndexed.Inc(1)
	}

	t.logger.Info().Dur("dur", time.Since(startTS)).Msg("indexed")
}

func (t *FB2IndexTask) indexFB2File(data io.Reader) bool {
	fb2File, err := fb2parser.UnmarshalStream(data)
	if err != nil {
		t.logger.Error().Err(err).Msg("parsing fb2 file")
		return false
	}

	logger := t.logger.With().Str("fb2-title", fb2File.Description.TitleInfo.BookTitle).Logger()

	doc := entities.NewBookIndex(fb2File)
	doc.Src = path.Join(t.fb2Dir, t.fb2File.Name())
	doc.SizeUncompressed = uint64(t.fb2File.Size())

	if err := t.index.Index(doc.ID, doc); err != nil {
		logger.Error().Err(err).Msg("indexing fb2")
		return false
	}

	return true
}
