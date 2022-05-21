package tasks

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
	libName     string
	itemPath    string
	itemPathRaw string
	repo        entities.IBooksIndexRepo
	logger      zerolog.Logger
	bar         *mpb.Bar
	cntTotal    *entities.CntAtomic32
	cntIndexed  *entities.CntAtomic32
	wg          *sync.WaitGroup
}

func NewFB2IndexTask(
	libName string,
	itemPath string,
	itemPathRaw string,
	repo entities.IBooksIndexRepo,
	logger zerolog.Logger,
	bar *mpb.Bar,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	wg *sync.WaitGroup,
) *FB2IndexTask {
	return &FB2IndexTask{
		libName:     libName,
		itemPath:    itemPath,
		itemPathRaw: itemPathRaw,
		repo:        repo,
		cntTotal:    cntTotal,
		cntIndexed:  cntIndexed,
		wg:          wg,
		logger:      logger.With().Str("task", "fb2_index").Logger(),
	}
}

func (t *FB2IndexTask) GetID() string {
	return fmt.Sprintf("fb2_index [%s] %s", t.libName, t.itemPath)
}

func (t *FB2IndexTask) Do() {
	defer func() {
		t.cntTotal.Inc(1)
		t.wg.Done()
	}()

	file, err := os.Open(t.itemPathRaw)
	if err != nil {
		t.logger.Error().Err(err).Msg("open item")
		return
	}
	defer file.Close()

	finfo, _ := file.Stat()
	if t.bar != nil {
		defer t.bar.IncrInt64(finfo.Size())
	}

	var fb2File fb2parser.FB2File
	if err = fb2parser.UnmarshalStream(file, &fb2File); err != nil {
		t.logger.Error().Err(err).Msg("parse item")
		return
	}

	doc := entities.NewBookIndex(&fb2File)
	doc.SizeUncompressed = uint64(finfo.Size())
	doc.LibName = t.libName
	doc.Src = t.itemPath
	doc.ID = entities.GenerateID([]string{doc.ISBN, doc.Lang, fmt.Sprint(doc.Year)},
		strings.Split(doc.Titles, ";"),
		strings.Split(strings.ReplaceAll(doc.Authors, ",", ";"), ";"),
	)

	if err = t.repo.SaveBook(doc); err != nil {
		t.logger.Error().Err(err).
			Str("bookname", fb2File.Description.TitleInfo.BookTitle).
			Msg("indexing")
		return
	}

	t.cntIndexed.Inc(1)
	t.logger.Debug().Msg("indexed")
}
