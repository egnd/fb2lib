package indexing

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/fb2parser"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
)

type FB2IndexTask struct {
	pathStr    string
	lib        entities.Library
	repo       entities.IBooksIndexRepo
	logger     zerolog.Logger
	bar        *mpb.Bar
	cntTotal   *entities.CntAtomic32
	cntIndexed *entities.CntAtomic32
	wg         *sync.WaitGroup
}

func NewFB2IndexTask(
	pathStr string,
	lib entities.Library,
	repo entities.IBooksIndexRepo,
	logger zerolog.Logger,
	bar *mpb.Bar,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	wg *sync.WaitGroup,
) *FB2IndexTask {
	return &FB2IndexTask{
		pathStr:    pathStr,
		lib:        lib,
		repo:       repo,
		logger:     logger.With().Str("task", "fb2index").Logger(),
		cntTotal:   cntTotal,
		cntIndexed: cntIndexed,
		wg:         wg,
	}
}

func (t *FB2IndexTask) GetID() string {
	return fmt.Sprintf("fb2index %s", t.pathStr)
}

func (t *FB2IndexTask) Do() {
	defer func() {
		t.cntTotal.Inc(1)
		t.wg.Done()
	}()

	file, err := os.Open(t.pathStr)
	if err != nil {
		t.logger.Error().Err(err).Msg("open fb2 file")
		return
	}
	defer file.Close()

	finfo, _ := file.Stat()
	if t.bar != nil {
		defer t.bar.IncrInt64(finfo.Size())
	}

	fb2File, err := fb2parser.UnmarshalStream(file)
	if err != nil {
		t.logger.Error().Err(err).Msg("parse fb2 file")
		return
	}

	doc := entities.NewBookIndex(fb2File)
	doc.SizeUncompressed = uint64(finfo.Size())
	doc.LibName = t.lib.Name
	doc.Src = strings.TrimPrefix(t.pathStr, t.lib.BooksDir)
	doc.ID = entities.GenerateID([]string{doc.ISBN, doc.Lang, fmt.Sprint(doc.Year)},
		strings.Split(doc.Titles, ";"),
		strings.Split(strings.ReplaceAll(doc.Authors, ",", ";"), ";"),
	)

	// doc.Offset = t.book.Archive.Offset
	// doc.SizeCompressed = t.book.Archive.Size
	// doc.Src = path.Join(t.book.Archive.Path, t.book.Path)

	if err = t.repo.SaveBook(doc); err != nil {
		t.logger.Error().Err(err).
			Str("bookname", fb2File.Description.TitleInfo.BookTitle).
			Msg("index fb2 file")
	}

	t.cntIndexed.Inc(1)
	t.logger.Debug().Msg("indexed")
}
