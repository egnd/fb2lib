package tasks

import (
	"fmt"
	"os"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/fb2parser"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
)

type FB2IndexTask struct {
	libName     string
	itemPath    string
	itemPathRaw string
	repo        entities.IBooksInfoRepo
	logger      zerolog.Logger
	bar         *mpb.Bar
	cntTotal    *entities.CntAtomic32
	cntIndexed  *entities.CntAtomic32
	wg          *sync.WaitGroup
	repoMarks   entities.ILibMarksRepo
}

func NewFB2IndexTask(
	libName string,
	itemPath string,
	itemPathRaw string,
	repo entities.IBooksInfoRepo,
	logger zerolog.Logger,
	bar *mpb.Bar,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	wg *sync.WaitGroup,
	repoMarks entities.ILibMarksRepo,
) *FB2IndexTask {
	return &FB2IndexTask{
		libName:     libName,
		itemPath:    itemPath,
		itemPathRaw: itemPathRaw,
		repo:        repo,
		cntTotal:    cntTotal,
		cntIndexed:  cntIndexed,
		wg:          wg,
		repoMarks:   repoMarks,
		logger:      logger.With().Str("task", "fb2_index").Logger(),
	}
}

func (t *FB2IndexTask) GetID() string {
	return fmt.Sprintf("fb2_index [%s] %s", t.libName, t.itemPath)
}

func (t *FB2IndexTask) Do() {
	defer t.wg.Done()

	finfo, err := os.Stat(t.itemPathRaw)
	if err != nil {
		t.logger.Error().Err(err).Msg("stat item")
		return
	}

	if t.bar != nil {
		defer t.bar.IncrInt64(finfo.Size())
	}

	if t.repoMarks.MarkExists(t.libName + t.itemPath) {
		t.logger.Info().Msg("already indexed")
		return
	}

	defer t.cntTotal.Inc(1)

	file, err := os.Open(t.itemPathRaw)
	if err != nil {
		t.logger.Error().Err(err).Msg("open item")
		return
	}
	defer file.Close()

	var fb2File fb2parser.FB2File
	if err = fb2parser.UnmarshalStream(file, &fb2File); err != nil {
		t.logger.Error().Err(err).Msg("parse item")
		return
	}

	if err = t.repo.SaveBook(entities.BookInfo{
		Index:            entities.NewBookIndex(&fb2File),
		SizeUncompressed: uint64(finfo.Size()),
		LibName:          t.libName,
		Src:              t.itemPath,
	}); err != nil {
		t.logger.Error().Err(err).
			Str("bookname", fb2File.Description.TitleInfo.BookTitle).
			Msg("indexing")
		return
	}

	if err := t.repoMarks.AddMark(t.libName + t.itemPath); err != nil {
		t.logger.Error().Err(err).Msg("save item index")
	}

	t.cntIndexed.Inc(1)
	t.logger.Debug().Msg("indexed")
}
