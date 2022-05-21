package tasks

import (
	"archive/zip"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-wpool/v2/interfaces"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"
)

type ZipIndexTask struct {
	pathStr    string
	lib        entities.Library
	repo       entities.IBooksIndexRepo
	logger     zerolog.Logger
	bar        *mpb.Bar
	cntTotal   *entities.CntAtomic32
	cntIndexed *entities.CntAtomic32
	wg         *sync.WaitGroup
	pool       interfaces.Pool
}

func NewZipIndexTask(
	pathStr string,
	lib entities.Library,
	repo entities.IBooksIndexRepo,
	logger zerolog.Logger,
	bar *mpb.Bar,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	wg *sync.WaitGroup,
	pool interfaces.Pool,
) *ZipIndexTask {
	return &ZipIndexTask{
		pathStr:    pathStr,
		lib:        lib,
		repo:       repo,
		bar:        bar,
		cntTotal:   cntTotal,
		cntIndexed: cntIndexed,
		wg:         wg,
		pool:       pool,
		logger:     logger.With().Str("task", "zip_index").Logger(),
	}
}

func (t *ZipIndexTask) GetID() string {
	return fmt.Sprintf("zip_index %s", t.pathStr)
}

func (t *ZipIndexTask) Do() {
	if t.alreadyIndexed() {
		return
	}

	archive, err := os.Open(t.pathStr)
	if err != nil {
		t.logger.Error().Err(err).Msg("open zip file")
		return
	}

	// defer archive.Close() // @TODO:

	archiveReader, err := zip.OpenReader(t.pathStr)
	if err != nil {
		t.logger.Error().Err(err).Msg("read zip file")
		return
	}

	// defer archiveReader.Close() // @TODO:

	for _, innerFile := range archiveReader.File {
		t.wg.Add(1)
		logger := t.logger.With().Str("libsubitem", innerFile.Name).Logger()

		if err := t.pool.AddTask(NewZipFB2IndexTask(t.lib.Name,
			path.Join(strings.TrimPrefix(t.pathStr, t.lib.BooksDir), innerFile.Name),
			archive, innerFile, t.repo, logger, t.bar, t.cntTotal, t.cntIndexed, t.wg,
		)); err != nil {
			logger.Error().Err(err).Str("libsubitem", innerFile.Name).Msg("handle subitem")
		}
	}

	if err := t.memorize(); err != nil {
		t.logger.Error().Err(err).Msg("handle zip file")
	}
}

func (t *ZipIndexTask) alreadyIndexed() bool {
	return false // @TODO:
}

func (t *ZipIndexTask) memorize() error {
	return nil // @TODO:
}
