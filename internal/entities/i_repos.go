package entities

import (
	"github.com/egnd/fb2lib/pkg/pagination"
)

type IBooksIndexRepo interface {
	SearchAll(string, pagination.IPager) ([]BookIndex, error)
	SearchByAuthor(string, pagination.IPager) ([]BookIndex, error)
	SearchBySequence(string, pagination.IPager) ([]BookIndex, error)
	GetBook(string) (BookIndex, error)
}

type IBooksDataRepo interface {
	GetFor(BookIndex) (FB2Book, error)
}
