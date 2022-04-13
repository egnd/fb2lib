package repos

import (
	"context"
	"errors"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/pkg/pagination"
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
	ctx context.Context, strQuery string, pager pagination.IPager,
) (res []entities.BookIndex, err error) {
	strQuery = strings.TrimSpace(strings.ToLower(strQuery))

	var q query.Query
	if strQuery == "" || strQuery == "*" {
		q = bleve.NewMatchAllQuery()
	} else {
		q = r.getCompositeQuery(strQuery)
	}

	searchReq := bleve.NewSearchRequestOptions(q, pager.GetPageSize(), pager.GetOffset(), false)
	searchReq.Fields = []string{"*"}
	searchReq.Highlight = bleve.NewHighlightWithStyle("html")
	searchReq.Sort = append(searchReq.Sort, &search.SortField{
		Field:   "Date",
		Desc:    true,
		Type:    search.SortFieldAsString,
		Missing: search.SortFieldMissingLast,
	})

	var searchResults *bleve.SearchResult
	if searchResults, err = r.index.Search(searchReq); err != nil {
		return
	}

	pager.SetTotal(searchResults.Total)

	for _, item := range searchResults.Hits {
		book := entities.BookIndex{
			ID:        item.ID,
			ISBN:      item.Fields["ISBN"].(string),
			Titles:    item.Fields["Titles"].(string),
			Authors:   item.Fields["Authors"].(string),
			Sequences: item.Fields["Sequences"].(string),
			Date:      item.Fields["Date"].(string),
			Publisher: item.Fields["Publisher"].(string),
			// Genres:           item.Fields["Genres"].(string), //@TODO: uncomment
			SizeCompressed:   item.Fields["SizeCompressed"].(float64),
			SizeUncompressed: item.Fields["SizeUncompressed"].(float64),
		}

		if item.Fields["Genres"] != nil { //@TODO: remove
			book.Genres = item.Fields["Genres"].(string)
		}

		if r.highlight {
			book = r.highlightItem(item.Fragments, book)
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
	if bookID == "" {
		err = errors.New("repo get book error: empty book id")

		return
	}

	searchReq := bleve.NewSearchRequestOptions(bleve.NewDocIDQuery([]string{bookID}), 1, 0, false)
	searchReq.Fields = []string{"*"}

	var searchResults *bleve.SearchResult
	if searchResults, err = r.index.Search(searchReq); err != nil {
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

	if searchResults.Hits[0].Fields["Genres"] != nil { //@TODO: remove
		res.Genres = searchResults.Hits[0].Fields["Genres"].(string)
	}

	res.Src = searchResults.Hits[0].Fields["Src"].(string)
	res.Offset = searchResults.Hits[0].Fields["Offset"].(float64)
	res.SizeCompressed = searchResults.Hits[0].Fields["SizeCompressed"].(float64)
	res.SizeUncompressed = searchResults.Hits[0].Fields["SizeUncompressed"].(float64)

	return
}
