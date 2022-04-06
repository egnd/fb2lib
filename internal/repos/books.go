package repos

import (
	"context"
	"errors"
	"strings"

	"github.com/astaxie/beego/utils/pagination"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

type BooksBleveRepo struct {
	highlight bool
	index     entities.ISearchIndex
	logger    zerolog.Logger
}

func NewBooksBleve(
	highlight bool, index entities.ISearchIndex, logger zerolog.Logger,
) *BooksBleveRepo {
	return &BooksBleveRepo{
		highlight: highlight,
		index:     index,
		logger:    logger,
	}
}

func (r *BooksBleveRepo) GetBooks(
	ctx context.Context, strQuery string, pager *pagination.Paginator,
) (res []entities.BookIndex, err error) {
	strQuery = strings.TrimSpace(strings.ToLower(strQuery))

	var q query.Query
	if strQuery == "" || strQuery == "*" {
		q = bleve.NewMatchAllQuery()
	} else {
		q = r.getCompositeQuery(strQuery)
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

		book := entities.BookIndex{
			ID:        searchResults.Hits[i].ID,
			ISBN:      searchResults.Hits[i].Fields["ISBN"].(string),
			Titles:    searchResults.Hits[i].Fields["Titles"].(string),
			Authors:   searchResults.Hits[i].Fields["Authors"].(string),
			Sequences: searchResults.Hits[i].Fields["Sequences"].(string),
			Date:      searchResults.Hits[i].Fields["Date"].(string),
			Publisher: searchResults.Hits[i].Fields["Publisher"].(string),
		}

		if r.highlight {
			book = r.highlightItem(searchResults.Hits[i].Fragments, book)
		}

		res = append(res, book)
	}

	return
}

func (r *BooksBleveRepo) getCompositeQuery(strQuery string) query.Query {
	queries := make([]query.Query, 0, 4)

	if strings.Contains(strQuery, " ") {
		queries = append(queries, bleve.NewMatchPhraseQuery(strQuery)) // phrase match
	} else {
		queries = append(queries, bleve.NewTermQuery(strQuery)) // exact word match
	}

	if strings.Contains(strQuery, "*") {
		queries = append(queries, bleve.NewWildcardQuery(strQuery)) // wildcard search syntax (*)
	} else {
		queries = append(queries, bleve.NewQueryStringQuery(strQuery)) // extended search syntax https://blevesearch.com/docs/Query-String-Query/
	}

	return bleve.NewDisjunctionQuery(queries...)
}

func (r *BooksBleveRepo) highlightItem(fragments search.FieldFragmentMap, book entities.BookIndex) entities.BookIndex {
	if vals, ok := fragments["Titles"]; ok && len(vals) > 0 {
		book.Titles = vals[0]
	}

	if vals, ok := fragments["Authors"]; ok && len(vals) > 0 {
		book.Authors = vals[0]
	}

	if vals, ok := fragments["Sequences"]; ok && len(vals) > 0 {
		book.Sequences = vals[0]
	}

	if vals, ok := fragments["Publisher"]; ok && len(vals) > 0 {
		book.Publisher = vals[0]
	}

	return book
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
