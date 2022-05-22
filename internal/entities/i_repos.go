package entities

import (
	"io"

	"github.com/egnd/fb2lib/pkg/pagination"
)

type IBooksInfoRepo interface {
	io.Closer
	SearchAll(string, pagination.IPager) ([]BookInfo, error)
	SearchByAuthor(string, pagination.IPager) ([]BookInfo, error)
	SearchBySequence(string, pagination.IPager) ([]BookInfo, error)
	GetBook(string) (BookInfo, error)
	SaveBook(BookInfo) error
}

type IBooksLibraryRepo interface {
	GetFB2(BookInfo) (FB2Book, error)
}

type ILibMarksRepo interface {
	MarkExists(string) bool
	AddMark(string) error
}
