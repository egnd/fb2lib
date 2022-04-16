package tasks

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/factories"
	"gitlab.com/egnd/bookshelf/pkg/fb2parser"
	"gitlab.com/egnd/bookshelf/pkg/library"
)

type ZIPFB2IndexTask struct {
	rewriteIndex bool
	useXMLMarsh  bool
	archiveDir   string
	indexDir     string
	archiveFile  os.FileInfo
	logger       zerolog.Logger
	wg           *sync.WaitGroup
	cntTotal     *entities.CntAtomic32
	cntIndexed   *entities.CntAtomic32
	itemsTotal   uint32
	itemsIndexed uint32
	index        entities.ISearchIndex
	barContainer *mpb.Progress
	bar          *mpb.Bar
	totalBar     *mpb.Bar
}

func NewZIPFB2IndexTask(
	archiveFile os.FileInfo,
	archiveDir string,
	indexDir string,
	rewriteIndex bool,
	useXMLMarsh bool,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	logger zerolog.Logger,
	wg *sync.WaitGroup,
	barContainer *mpb.Progress,
	totalBar *mpb.Bar,
) *ZIPFB2IndexTask {
	return &ZIPFB2IndexTask{
		rewriteIndex: rewriteIndex,
		archiveDir:   archiveDir,
		indexDir:     indexDir,
		archiveFile:  archiveFile,
		logger:       logger,
		wg:           wg,
		cntTotal:     cntTotal,
		cntIndexed:   cntIndexed,
		barContainer: barContainer,
		totalBar:     totalBar,
		useXMLMarsh:  useXMLMarsh,
	}
}

func (t *ZIPFB2IndexTask) Do(context.Context) {
	defer func() {
		if t.totalBar != nil {
			t.totalBar.IncrInt64(t.archiveFile.Size())
		}

		t.cntTotal.Inc(t.itemsTotal)
		t.cntIndexed.Inc(t.itemsIndexed)
		t.wg.Done()
	}()

	startTS := time.Now()

	var err error
	t.index, err = factories.NewTmpIndex(
		t.archiveFile, t.indexDir, t.rewriteIndex, entities.NewBookIndexMapping(),
	)
	if err != nil {
		t.logger.Warn().Err(err).Msg("init index")
		return
	}
	defer t.index.Close()

	if t.barContainer != nil {
		t.bar = t.initBar()
		defer t.bar.Abort(true)
	}

	if err := library.NewZipItemIterator(path.Join(t.archiveDir, t.archiveFile.Name()), t.logger).
		IterateItems(t.handleArchiveItem); err != nil {
		t.logger.Error().Err(err).Msg("iterate over archive")
		return
	}

	if err = factories.SaveTmpIndex(t.index); err != nil {
		t.logger.Error().Err(err).Msg("save index")
		return
	}

	t.logger.Info().
		Uint32("total", t.itemsTotal).
		Uint32("indexed", t.itemsIndexed).
		Dur("dur", time.Since(startTS)).
		Msg("indexed")

}

func (t *ZIPFB2IndexTask) initBar() *mpb.Bar {
	return t.barContainer.AddBar(t.archiveFile.Size(),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name("thread: "),
			decor.Name(t.archiveFile.Name()),
		),
		mpb.AppendDecorators(
			decor.AverageETA(decor.ET_STYLE_GO),
			decor.Name(" - "),
			decor.AverageSpeed(decor.UnitKB, "% .2f"),
			decor.Name(" - "),
			decor.CountersKibiByte("% .2f/% .2f"),
		),
	)
}

func (t *ZIPFB2IndexTask) handleArchiveItem(zipItem *zip.File, data io.Reader, offset, num int64, logger zerolog.Logger) error {
	if t.bar != nil {
		defer t.bar.IncrInt64(int64(zipItem.CompressedSize64))
	}

	switch path.Ext(zipItem.Name) {
	case ".fb2":
		t.itemsTotal++

		var err error
		var fb2File *fb2parser.FB2File

		if t.useXMLMarsh {
			fb2File, err = fb2parser.UnmarshalFB2Stream(data)
		} else {
			fb2File, err = fb2parser.ParseFB2Stream(data)
		}

		if err != nil {
			logger.Error().Err(err).Msg("parsing fb2 file")
			return nil
		}

		logger = logger.With().Str("fb2-title", fb2File.Description.TitleInfo.BookTitle).Logger()

		doc := entities.NewBookIndex(fb2File)
		doc.Src = path.Join(t.archiveDir, t.archiveFile.Name(), zipItem.Name)
		doc.Offset = float64(offset)
		doc.SizeCompressed = float64(zipItem.CompressedSize64)
		doc.SizeUncompressed = float64(zipItem.UncompressedSize64)

		if err := t.index.Index(doc.ID, doc); err != nil {
			logger.Error().Err(err).Msg("indexing fb2")
			return nil
		}

		t.itemsIndexed++

		return nil
	default:
		logger.Warn().Msg("invalid archive item")

		return nil
	}
}
