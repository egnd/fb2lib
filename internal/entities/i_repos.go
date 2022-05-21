package entities

import (
	"io"

	"github.com/egnd/fb2lib/pkg/pagination"
)

type IBooksIndexRepo interface {
	io.Closer
	SearchAll(string, pagination.IPager) ([]BookIndex, error)
	SearchByAuthor(string, pagination.IPager) ([]BookIndex, error)
	SearchBySequence(string, pagination.IPager) ([]BookIndex, error)
	GetBook(string) (BookIndex, error)
	SaveBook(BookIndex) error
}

type IBooksDataRepo interface {
	GetFB2(BookIndex) (FB2Book, error)
}
