package entities

import (
	"gitlab.com/egnd/bookshelf/pkg/pagination"
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
