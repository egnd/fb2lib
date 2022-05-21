package tasks

import (
	"archive/zip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/fb2parser"
	"github.com/egnd/fb2lib/pkg/library"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

type ZIPFB2IndexTask struct {
	itemsTotal    uint32
	itemsIndexed  uint32
	batchSize     int
	archiveDir    string
	markPath      string
	batchChan     chan entities.BookIndex
	batchStopChan chan struct{}
	archive       os.FileInfo
	lib           entities.Library
	logger        zerolog.Logger
	wg            *sync.WaitGroup
	cntTotal      *entities.CntAtomic32
	cntIndexed    *entities.CntAtomic32
	barContainer  *mpb.Progress
	bar           *mpb.Bar
	totalBar      *mpb.Bar
	index         entities.ISearchIndex
}

func NewZIPFB2IndexTask(
	archive os.FileInfo,
	archiveDir string,
	batchSize int,
	lib entities.Library,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	logger zerolog.Logger,
	wg *sync.WaitGroup,
	barContainer *mpb.Progress,
	totalBar *mpb.Bar,
	index entities.ISearchIndex,
) *ZIPFB2IndexTask {
	hasher := md5.New()
	hasher.Write([]byte(lib.Name))
	hasher.Write([]byte(archiveDir))
	hasher.Write([]byte(archive.Name()))

	return &ZIPFB2IndexTask{
		markPath:     path.Join(lib.IndexDir, hex.EncodeToString(hasher.Sum(nil))+".mark"),
		archiveDir:   archiveDir,
		archive:      archive,
		logger:       logger,
		wg:           wg,
		cntTotal:     cntTotal,
		cntIndexed:   cntIndexed,
		barContainer: barContainer,
		totalBar:     totalBar,
		batchSize:    batchSize,
		index:        index,
		lib:          lib,
	}
}

func (t *ZIPFB2IndexTask) Do(context.Context) {
	defer func() {
		if t.totalBar != nil {
			t.totalBar.IncrInt64(t.archive.Size())
		}

		t.cntTotal.Inc(t.itemsTotal)
		t.cntIndexed.Inc(t.itemsIndexed)
		t.wg.Done()
	}()

	if _, err := os.Stat(t.markPath); err == nil {
		t.logger.Info().
			Str("path", path.Join(t.archiveDir, t.archive.Name())).
			Msg("archive already indexed")

		return
	}

	startTS := time.Now()

	if t.barContainer != nil {
		t.bar = t.initBar()
		defer t.bar.Abort(true)
	}

	t.batchChan = make(chan entities.BookIndex)
	defer close(t.batchChan)

	t.batchStopChan = make(chan struct{})
	defer close(t.batchStopChan)

	var batchWg sync.WaitGroup
	batchWg.Add(1)

	go t.runBatcher(&batchWg)

	if err := library.NewZipItemIterator(
		path.Join(t.lib.BooksDir, t.archiveDir, t.archive.Name()), t.logger,
	).IterateItems(t.handleArchiveItem); err != nil {
		t.logger.Error().Err(err).Msg("iterate over archive")
		t.batchStopChan <- struct{}{}
		batchWg.Wait()

		return
	}

	t.batchStopChan <- struct{}{}
	batchWg.Wait()

	t.logger.Info().
		Uint32("total", t.itemsTotal).
		Uint32("indexed", t.itemsIndexed).
		Dur("dur", time.Since(startTS)).
		Msg("indexed")
}

func (t *ZIPFB2IndexTask) initBar() *mpb.Bar {
	return t.barContainer.AddBar(t.archive.Size(),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name("thread: "),
			decor.Name(t.archive.Name()),
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

		fb2File, err := fb2parser.UnmarshalStream(data)
		if err != nil {
			logger.Error().Err(err).Msg("parsing fb2 file")
			return nil
		}

		logger = logger.With().Str("fb2-title", fb2File.Description.TitleInfo.BookTitle).Logger()

		doc := entities.NewBookIndex(fb2File)
		doc.LibName = t.lib.Name
		doc.Src = path.Join(t.archiveDir, t.archive.Name(), zipItem.Name)
		doc.Offset = uint64(offset)
		doc.SizeCompressed = zipItem.CompressedSize64
		doc.SizeUncompressed = zipItem.UncompressedSize64
		doc.ID = entities.GenerateID([]string{doc.ISBN, doc.Lang, fmt.Sprint(doc.Year)},
			strings.Split(doc.Titles, ";"),
			strings.Split(strings.ReplaceAll(doc.Authors, ",", ";"), ";"),
		)

		t.batchChan <- doc

		return nil
	default:
		logger.Warn().Msg("invalid archive item")

		return nil
	}
}

func (t *ZIPFB2IndexTask) runBatcher(wg *sync.WaitGroup) {
	defer wg.Done()

	batch := t.index.NewBatch()

	defer func() {
		if batch.Size() > 0 {
			if err := t.index.Batch(batch); err != nil {
				t.logger.Error().Err(err).Msg("exec index batch last")
			} else {
				t.itemsIndexed += uint32(batch.Size())
			}
		}

		os.WriteFile(t.markPath, []byte{}, 0644)
	}()

	for {
		select {
		case <-t.batchStopChan:
			return
		case doc := <-t.batchChan:
			if err := batch.Index(doc.ID, doc); err != nil {
				t.logger.Error().Err(err).Msg("add book to index batch")
				continue
			}

			if batch.Size() < t.batchSize {
				continue
			}

			if err := t.index.Batch(batch); err != nil {
				t.logger.Error().Err(err).Msg("exec index batch")
			} else {
				t.itemsIndexed += uint32(batch.Size())
			}

			batch.Reset()
		}
	}
}
