package tasks

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/egnd/go-fb2parse"
	"github.com/rs/zerolog"
	"github.com/vbauerster/mpb/v7"

	"github.com/egnd/fb2lib/internal/entities"
)

type ZipFB2IndexTask struct {
	libName    string
	itemPath   string
	encoder    entities.LibEncodeType
	parent     *os.File
	file       *zip.File
	repo       entities.IBooksInfoRepo
	logger     zerolog.Logger
	bar        *mpb.Bar
	cntTotal   *entities.CntAtomic32
	cntIndexed *entities.CntAtomic32
	wg         *sync.WaitGroup
}

func NewZipFB2IndexTask(
	libName string,
	encoder entities.LibEncodeType,
	itemPath string,
	parent *os.File,
	file *zip.File,
	repo entities.IBooksInfoRepo,
	logger zerolog.Logger,
	bar *mpb.Bar,
	cntTotal *entities.CntAtomic32,
	cntIndexed *entities.CntAtomic32,
	wg *sync.WaitGroup,
) *ZipFB2IndexTask {
	return &ZipFB2IndexTask{
		libName:    libName,
		itemPath:   itemPath,
		encoder:    encoder,
		parent:     parent,
		file:       file,
		repo:       repo,
		bar:        bar,
		cntTotal:   cntTotal,
		cntIndexed: cntIndexed,
		wg:         wg,
		logger:     logger.With().Str("task", "zip_fb2_index").Logger(),
	}
}

func (t *ZipFB2IndexTask) GetID() string {
	return fmt.Sprintf("zip_fb2_index [%s] %s", t.libName, t.itemPath)
}

func (t *ZipFB2IndexTask) Do() {
	defer func() {
		if t.bar != nil {
			t.bar.IncrInt64(int64(t.file.CompressedSize64))
		}

		t.cntTotal.Inc(1)
		t.wg.Done()
	}()

	offset, err := t.file.DataOffset()
	if err != nil {
		t.logger.Error().Err(err).Msg("get offset")
		return
	}

	var data io.ReadCloser
	switch t.file.Method {
	case zip.Deflate:
		data = flate.NewReader(io.NewSectionReader(t.parent, offset, int64(t.file.CompressedSize64)))
	case zip.Store:
		data = io.NopCloser(io.NewSectionReader(t.parent, offset, int64(t.file.CompressedSize64)))
	default:
		t.logger.Warn().Uint16("method", t.file.Method).Msg("define compression")
		return
	}

	defer data.Close()

	var fb2File fb2parse.FB2File
	if fb2File, err = entities.ParseFB2(data, t.encoder,
		SkipFB2Binaries, SkipFB2DocInfo, SkipFB2CustomInfo, SkipFB2Cover); err != nil {
		t.logger.Error().Err(err).Msg("parse zipped fb2 file")
		return
	}

	if err = t.repo.SaveBook(entities.BookInfo{
		Index:          entities.NewFB2Index(&fb2File),
		Offset:         uint64(offset),
		SizeCompressed: uint64(t.file.CompressedSize64),
		Size:           uint64(t.file.UncompressedSize64),
		LibName:        t.libName,
		Src:            t.itemPath,
	}); err != nil {
		t.logger.Error().Err(err).
			Str("bookname", entities.GetFirstStr(fb2File.Description[0].TitleInfo[0].BookTitle)).
			Msg("index zipped fb2 file")
		return
	}

	t.cntIndexed.Inc(1)
	t.logger.Debug().Msg("indexed")
}
