package repos

import (
	"compress/flate"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/tasks"
	"github.com/egnd/go-xmlparse"
	"github.com/egnd/go-xmlparse/fb2"
	"github.com/rs/zerolog"

	"github.com/egnd/fb2lib/internal/entities"
)

type LibraryFs struct {
	libs     entities.Libraries
	logger   zerolog.Logger
	executor pipeline.Dispatcher
}

func NewLibraryFs(libs entities.Libraries, executor pipeline.Dispatcher, logger zerolog.Logger) *LibraryFs {
	return &LibraryFs{
		libs:     libs,
		logger:   logger,
		executor: executor,
	}
}

func (r *LibraryFs) AppendFB2Book(book *entities.Book) error {
	fb2File, err := r.readFB2(book, getBookCoverRule(book))
	if err != nil {
		return err
	}

	for _, bin := range fb2File.Binary {
		if bin.ID == book.Info.CoverID {
			book.Info.Cover = &bin
		}

		if book.OrigInfo != nil && bin.ID == book.OrigInfo.CoverID {
			book.OrigInfo.Cover = &bin
		}
	}

	return nil
}

func (r *LibraryFs) AppendFB2Books(books []entities.Book) {
	for k := range books {
		k := k
		r.executor.Push(tasks.NewFunc("", func() error {
			return r.AppendFB2Book(&books[k])
		}))
	}

	r.executor.Wait()
}

func (r *LibraryFs) readFB2(book *entities.Book, rules ...xmlparse.Rule) (*fb2.File, error) {
	if book.Src == "" {
		return nil, fmt.Errorf("libsfs repo err: empty book src [%s]", book.ID)
	}

	lib, libExist := r.libs[book.Lib]
	if !libExist {
		return nil, fmt.Errorf("libsfs repo err: undefined lib name %s", book.Lib)
	}

	var fb2Stream io.ReadCloser
	if strings.Contains(book.Src, ".zip") {
		zipFile, err := os.Open(strings.Split(path.Join(lib.Dir, book.Src), ".zip")[0] + ".zip")
		if err != nil {
			return nil, err
		}
		defer zipFile.Close()

		fb2Stream = flate.NewReader(io.NewSectionReader(zipFile, int64(book.Offset), int64(book.SizeCompressed)))
	} else {
		var err error

		if fb2Stream, err = os.Open(path.Join(lib.Dir, book.Src)); err != nil {
			return nil, err
		}
	}

	defer fb2Stream.Close()

	res, err := entities.ParseFB2(fb2Stream, lib.Encoder, rules...)

	return &res, err
}

func getBookCoverRule(book *entities.Book) xmlparse.Rule {
	return func(next xmlparse.TokenHandler) xmlparse.TokenHandler {
		return func(obj interface{}, node xml.StartElement, r xmlparse.TokenReader) error {
			if _, ok := obj.(*fb2.File); !ok {
				return nil
			}

			if node.Name.Local != "binary" {
				return nil
			}

			var binID string
			for _, attr := range node.Attr {
				if attr.Name.Local == "id" {
					binID = attr.Value
					break
				}
			}

			if book.Info.CoverID == binID || (book.OrigInfo != nil && book.OrigInfo.CoverID == binID) {
				return next(obj, node, r)
			}

			return nil
		}
	}
}
