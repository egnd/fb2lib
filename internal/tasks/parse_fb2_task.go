package tasks

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/internal/repos"
)

type ErrSkipRule struct {
	book string
	err  error
}

func (e ErrSkipRule) Error() string {
	return fmt.Sprintf("skip: %s - %s", e.err, e.book)
}

type PushParseTask func(io.Reader) error

type ParseFB2Task struct {
	id      string
	data    io.Reader
	book    entities.Book
	encoder entities.LibEncodeType
	repo    *repos.BooksLevelBleve
	bar     *mpb.Bar
	rules   entities.IndexRules
}

func NewParseFB2Task(
	data io.Reader,
	book entities.Book,
	rules entities.IndexRules,
	encoder entities.LibEncodeType,
	repo *repos.BooksLevelBleve,
	bar *mpb.Bar,
) *ParseFB2Task {
	return &ParseFB2Task{
		id:      fmt.Sprintf("parse [%s] %s", book.Lib, book.Src),
		data:    data,
		book:    book,
		encoder: encoder,
		repo:    repo,
		bar:     bar,
		rules:   rules,
	}
}

func (t *ParseFB2Task) ID() string {
	return t.id
}

func (t *ParseFB2Task) Do() error {
	if t.bar != nil {
		if t.book.SizeCompressed > 0 {
			defer t.bar.IncrInt64(int64(t.book.SizeCompressed))
		} else {
			defer t.bar.IncrInt64(int64(t.book.Size))
		}
	}

	fb2File, err := entities.ParseFB2(t.data, t.encoder, SkipFB2DocInfo, SkipFB2CustomInfo, SkipFB2Binaries)

	if err != nil {
		return errors.Wrap(err, "parse fb2 error")
	}

	t.book.ReadFB2(&fb2File)

	if err := t.rules.Check(&t.book); err != nil {
		return &ErrSkipRule{t.book.Info.Title, err}
	}

	if err = t.repo.SaveBook(&t.book); err != nil {
		return errors.Wrap(err, "index fb2 error")
	}

	return nil
}
