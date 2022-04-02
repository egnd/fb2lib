package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"

	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/internal/factories"
	"gitlab.com/egnd/bookshelf/internal/library"
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

	booksIndex, err := factories.NewBooksIndex(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("init db")
	}

	var archives []fs.FileInfo
	if archives, err = ioutil.ReadDir(cfg.GetString("web.library_dir")); err != nil {
		return
	}

	var bar *progressbar.ProgressBar
	debug := cfg.GetBool("debug")
	totalArchives := len(archives)
	for archiveNum, archive := range archives {
		logger := logger.With().Str("archive", archive.Name()).Logger()
		switch filepath.Ext(archive.Name()) {
		case ".zip":
			if debug {
				if bar, err = factories.NewFileProgressBar(
					filepath.Join(cfg.GetString("web.library_dir"), archive.Name()),
					fmt.Sprintf("[%d/%d] indexing %s...", archiveNum+1, totalArchives, archive.Name()),
				); err != nil {
					logger.Error().Err(err).Msg("new file")
					continue
				}
			}

			var booksTotal, booksIndexed uint32
			startTs := time.Now()
			library.NewArchiveZip(filepath.Join(cfg.GetString("web.library_dir"), archive.Name())).
				Walk(logger, func(item *zip.File, data io.Reader, offset, num int64) error {
					if debug {
						bar.Add64(int64(item.CompressedSize64))
					}

					logger := logger.With().Str("item", item.Name).Logger()
					switch path.Ext(item.Name) {
					case ".fb2":
						booksTotal++
						fb2File, err := library.NewFB2FileFromReader(data)
						if err != nil {
							logger.Error().Err(err).Msg("parsing fb2 item")
							return nil
						}
						book := entities.NewBookIndexFrom(fb2File)
						book.ID = uuid.NewString()
						book.Archive = archive.Name()
						book.Offset = offset
						book.SizeCompressed = int64(item.CompressedSize64)
						if err := booksIndex.Index(book.ID, book); err != nil {
							logger.Error().Err(err).
								Str("title", book.Titles).Str("authors", book.Authors).
								Msg("book")
							return nil
						}
						booksIndexed++
					default:
						logger.Warn().Msg("invalid archive item")
					}
					return nil
				})

			if debug {
				bar.Finish()
			}

			logger.Info().
				Uint32("total", booksTotal).Uint32("indexed", booksIndexed).
				Dur("dur", time.Now().Sub(startTs)).
				Msg("archive processed")
		default:
			logger.Warn().Msg("invalid archive file")
		}
	}
	logger.Info().Msg("finished")
}
