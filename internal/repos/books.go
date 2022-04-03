package repos

import (
	"context"
	"errors"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

type BooksBleveRepo struct {
	index  bleve.Index
	logger zerolog.Logger
}

func NewBooksBleve(index bleve.Index, logger zerolog.Logger) *BooksBleveRepo {
	return &BooksBleveRepo{
		index:  index,
		logger: logger,
	}
}

func (r *BooksBleveRepo) GetBooks(ctx context.Context, strQuery string, pager *pagination.Paginator) (res []entities.BookIndex, err error) {
	var q query.Query

	if strQuery == "" {
		q = bleve.NewMatchAllQuery()
	} else {
		q = bleve.NewMatchQuery(strQuery)
	}

	cnt, _ := r.index.DocCount()
	search := bleve.NewSearchRequestOptions(q, int(cnt), 0, false)
	search.Fields = []string{"*"}
	search.Highlight = bleve.NewHighlightWithStyle("html")

	var searchResults *bleve.SearchResult
	if searchResults, err = r.index.Search(search); err != nil {
		return
	}

	if searchResults.Total == 0 {
		return
	}

	pager.SetNums(searchResults.Total)
	totalHits := len(searchResults.Hits)

	for i := pager.Offset(); i < pager.Offset()+pager.PerPageNums; i++ {
		if totalHits <= i {
			break
		}

		res = append(res, entities.BookIndex{
			ID:        searchResults.Hits[i].ID,
			ISBN:      searchResults.Hits[i].Fields["ISBN"].(string),
			Titles:    searchResults.Hits[i].Fields["Titles"].(string),
			Authors:   searchResults.Hits[i].Fields["Authors"].(string),
			Sequences: searchResults.Hits[i].Fields["Sequences"].(string),
			Date:      searchResults.Hits[i].Fields["Date"].(string),
			Publisher: searchResults.Hits[i].Fields["Publisher"].(string),
		})
	}

	return
}

func (r *BooksBleveRepo) GetBook(ctx context.Context, bookID string) (res entities.BookIndex, err error) {
	search := bleve.NewSearchRequestOptions(bleve.NewDocIDQuery([]string{bookID}), 1, 0, false)
	search.Fields = []string{"*"}

	var searchResults *bleve.SearchResult
	if searchResults, err = r.index.Search(search); err != nil {
		return
	}

	if searchResults.Total == 0 {
		err = errors.New("book not found")
		return
	}

	res.ID = searchResults.Hits[0].ID
	res.ISBN = searchResults.Hits[0].Fields["ISBN"].(string)
	res.Titles = searchResults.Hits[0].Fields["Titles"].(string)
	res.Authors = searchResults.Hits[0].Fields["Authors"].(string)
	res.Sequences = searchResults.Hits[0].Fields["Sequences"].(string)
	res.Date = searchResults.Hits[0].Fields["Date"].(string)
	res.Publisher = searchResults.Hits[0].Fields["Publisher"].(string)
	res.Src = searchResults.Hits[0].Fields["Src"].(string)
	res.Offset = uint64(searchResults.Hits[0].Fields["Offset"].(float64))
	res.Size = uint64(searchResults.Hits[0].Fields["Size"].(float64))

	return
}
