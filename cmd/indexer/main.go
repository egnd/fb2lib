package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"

	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/factories"
	"gitlab.com/egnd/bookshelf/pkg/fb2parser"
	"gitlab.com/egnd/bookshelf/pkg/library"
)

var appVersion = "debug"

func main() {
	showVersion := flag.Bool("version", false, "Show app version.")
	cfgPath := flag.String("config", "configs/app.yml", "Configuration file path.")
	cfgPrefix := flag.String("env-prefix", "BS", "Prefix for env variables.")
	flag.Parse()

	if *showVersion {
		fmt.Println(appVersion)
		return
	}

	cfg, err := factories.NewViperCfg(*cfgPath, *cfgPrefix)
	if err != nil {
		log.Fatal().Err(err).Msg("init config")
	}

	logger := factories.NewZerologLogger(cfg, os.Stderr)

	if err = os.RemoveAll(cfg.GetString("bleve.path")); err != nil && !os.IsNotExist(err) {
		logger.Fatal().Err(err).Msg("remove index")
	}

	var bar *progressbar.ProgressBar
	var booksTotal, booksIndexed uint32

	debug := cfg.GetBool("debug")
	startTS := time.Now()
	booksIndexDir := cfg.GetString("bleve.books_dir")

	sepBooksIndex, err := factories.NewBooksIndex("notzip", booksIndexDir)
	if err != nil {
		logger.Fatal().Err(err).Msg("init notzip index")
	}
	defer sepBooksIndex.Close()

	lib := library.NewLocalFSItems(cfg.GetString("web.library_dir"), []string{".zip", ".fb2"}, logger)

	if err = lib.IterateItems(
		LibItemHandler(debug, bar, &booksTotal, &booksIndexed, booksIndexDir, sepBooksIndex),
	); err != nil {
		logger.Error().Err(err).Msg("handle library items")
	}

	logger.Info().
		Uint32("total", booksTotal).
		Uint32("indexed", booksIndexed).
		Dur("dur", time.Now().Sub(startTS)).
		Msg("indexing finished")
}

func LibItemHandler(
	debug bool, bar *progressbar.ProgressBar, booksTotal *uint32, booksIndexed *uint32,
	booksIndexDir string, sepBooksIndex bleve.Index,
) library.ILibItemHandler {
	return func(libItemPath string, libItem os.FileInfo, num, total int, logger zerolog.Logger) error {
		switch filepath.Ext(libItem.Name()) {
		case ".zip":
			if debug {
				bar = progressbar.DefaultBytes(libItem.Size(),
					fmt.Sprintf("[%d/%d] indexing %s...", num, total, libItem.Name()),
				)
				defer bar.Close()
			}

			var zipItemsTotal, zipItemsIndexed uint32
			defer func() {
				*booksTotal += zipItemsTotal
				*booksIndexed += zipItemsIndexed
			}()

			itemTS := time.Now()

			booksIndex, err := factories.NewBooksIndex(libItemPath, booksIndexDir)
			if err != nil {
				logger.Error().Err(err).Msg("init index")
				return nil
			}
			defer booksIndex.Close()

			if err := library.NewZipItemIterator(libItemPath, logger).IterateItems(ZipItemHandler(
				debug, bar, &zipItemsTotal, &zipItemsIndexed, libItemPath, booksIndex,
			)); err != nil {
				return err
			}

			if debug {
				bar.Finish()
			}

			logger.Info().Uint32("total", zipItemsTotal).Uint32("indexed", zipItemsIndexed).
				Dur("dur", time.Now().Sub(itemTS)).Msg("lib item indexed")

			return nil
		case ".fb2":
			fb2Data, err := os.Open(libItemPath)
			if err != nil {
				logger.Error().Err(err).Msg("open fb2 item")
				return nil
			}
			defer fb2Data.Close()

			*booksTotal++

			if !IndexFB2File(fb2Data, libItemPath, 0, uint64(libItem.Size()), logger, sepBooksIndex) {
				return nil
			}

			*booksIndexed++

			return nil
		default:
			return fmt.Errorf("lib item error: invalid item %s", libItem.Name())
		}
	}
}

func ZipItemHandler(
	debug bool, bar *progressbar.ProgressBar, zipItemsTotal *uint32, zipItemsIndexed *uint32,
	libItemPath string, booksIndex bleve.Index,
) library.IZipItemHandler {
	return func(zipItem *zip.File, data io.Reader, offset, num int64, logger zerolog.Logger) error {
		if debug {
			defer bar.Add64(int64(zipItem.CompressedSize64))
		}

		switch path.Ext(zipItem.Name) {
		case ".fb2":
			*zipItemsTotal++

			if !IndexFB2File(data, libItemPath, offset, zipItem.CompressedSize64, logger, booksIndex) {
				return nil
			}

			*zipItemsIndexed++
		default:
			logger.Warn().Msg("invalid archive item")
		}

		return nil
	}
}

func IndexFB2File(data io.Reader, srcPath string, offset int64, size uint64, logger zerolog.Logger, index bleve.Index) bool {
	fb2File, err := fb2parser.FB2FromReader(data)
	if err != nil {
		logger.Error().Err(err).Msg("parsing fb2 file")
		return false
	}

	logger = logger.With().Str("fb2-title", fb2File.Description.TitleInfo.BookTitle).Logger()

	doc := entities.NewBookIndex(fb2File)
	doc.ID = uuid.NewString()
	doc.Src = srcPath
	doc.Offset = uint64(offset)
	doc.Size = size

	if err := index.Index(doc.ID, doc); err != nil {
		logger.Error().Err(err).Msg("indexing fb2")
		return false
	}

	return true
}
