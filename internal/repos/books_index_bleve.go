package repos

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
	"gitlab.com/egnd/bookshelf/pkg/pagination"
)

var (
	regexpSpaces = regexp.MustCompile(`\s+`)
)

type BooksIndexBleve struct {
	highlight bool
	index     entities.ISearchIndex
	logger    zerolog.Logger
}

func NewBooksIndexBleve(
	highlight bool, index entities.ISearchIndex, logger zerolog.Logger,
) *BooksIndexBleve {
	return &BooksIndexBleve{
		highlight: highlight,
		index:     index,
		logger:    logger,
	}
}

func (r *BooksIndexBleve) SearchAll(strQuery string, pager pagination.IPager) ([]entities.BookIndex, error) {
	strQuery = strings.TrimSpace(strings.ToLower(strQuery))

	var q query.Query
	if strQuery == "" || strQuery == "*" {
		q = bleve.NewMatchAllQuery()
	} else {
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

		q = bleve.NewDisjunctionQuery(queries...)
	}

	searchReq := bleve.NewSearchRequestOptions(q, pager.GetPageSize(), pager.GetOffset(), false)
	searchReq.Highlight = bleve.NewHighlightWithStyle("html")
	searchReq.Sort = append(searchReq.Sort, &search.SortField{
		Field:   "Year",
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	})

	return r.getBooks(searchReq, pager)
}

func (r *BooksIndexBleve) SearchByAuthor(strQuery string, pager pagination.IPager) ([]entities.BookIndex, error) {
	strQuery = strings.TrimSpace(regexpSpaces.ReplaceAllString(strings.ToLower(strQuery), " "))

	var q query.Query
	if strQuery == "" {
		q = bleve.NewMatchAllQuery()
	} else {
		q = bleve.NewQueryStringQuery("+Authors:" + strings.ReplaceAll(strQuery, " ", " +Authors:"))
	}

	searchReq := bleve.NewSearchRequestOptions(q, pager.GetPageSize(), pager.GetOffset(), false)
	searchReq.Highlight = bleve.NewHighlightWithStyle("html")
	searchReq.Sort = append(searchReq.Sort, &search.SortField{
		Field:   "Year",
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	})

	return r.getBooks(searchReq, pager)
}

func (r *BooksIndexBleve) SearchBySequence(strQuery string, pager pagination.IPager) ([]entities.BookIndex, error) {
	strQuery = strings.TrimSpace(regexpSpaces.ReplaceAllString(strings.ToLower(strQuery), " "))

	var q query.Query
	if strQuery == "" {
		q = bleve.NewMatchAllQuery()
	} else {
		q = bleve.NewQueryStringQuery("+Sequences:" + strings.ReplaceAll(strQuery, " ", " +Sequences:"))
	}

	searchReq := bleve.NewSearchRequestOptions(q, pager.GetPageSize(), pager.GetOffset(), false)
	searchReq.Highlight = bleve.NewHighlightWithStyle("html")
	searchReq.Sort = append(searchReq.Sort, &search.SortField{
		Field:   "Year",
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	})

	return r.getBooks(searchReq, pager)
}

func (r *BooksIndexBleve) GetBook(bookID string) (entities.BookIndex, error) {
	if bookID == "" {
		return entities.BookIndex{}, errors.New("repo get book error: empty book id")
	}

	searchReq := bleve.NewSearchRequestOptions(
		bleve.NewDocIDQuery([]string{bookID}), 1, 0, false,
	)

	items, err := r.getBooks(searchReq, nil)
	if err != nil {
		return entities.BookIndex{}, err
	}

	if len(items) == 0 {
		return entities.BookIndex{}, fmt.Errorf("repo getbook error: %s not found", bookID)
	}

	return items[0], nil
}

func (r *BooksIndexBleve) highlightItem(fragments search.FieldFragmentMap, book entities.BookIndex) entities.BookIndex {
	if vals, ok := fragments["ISBN"]; ok && len(vals) > 0 {
		book.ISBN = vals[0]
	}

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

	if vals, ok := fragments["Date"]; ok && len(vals) > 0 {
		book.Date = vals[0]
	}

	if vals, ok := fragments["Genres"]; ok && len(vals) > 0 {
		book.Genres = vals[0]
	}

	return book
}

func (r *BooksIndexBleve) getBooks(
	searchReq *bleve.SearchRequest, pager pagination.IPager,
) ([]entities.BookIndex, error) {
	searchReq.Fields = []string{"*"}

	searchResults, err := r.index.Search(searchReq)
	if err != nil {
		return nil, err
	}

	if pager != nil {
		pager.SetTotal(searchResults.Total)
	}

	res := make([]entities.BookIndex, 0, len(searchResults.Hits))

	for _, item := range searchResults.Hits {
		book := entities.BookIndex{
			Offset:           uint64(item.Fields["Offset"].(float64)),
			SizeCompressed:   uint64(item.Fields["SizeCompressed"].(float64)),
			SizeUncompressed: uint64(item.Fields["SizeUncompressed"].(float64)),
			Lang:             item.Fields["Lang"].(string),
			Src:              item.Fields["Src"].(string),
			ID:               item.ID,
			ISBN:             item.Fields["ISBN"].(string),
			Titles:           item.Fields["Titles"].(string),
			Authors:          item.Fields["Authors"].(string),
			Sequences:        item.Fields["Sequences"].(string),
			Publisher:        item.Fields["Publisher"].(string),
			Date:             item.Fields["Date"].(string),
			Genres:           item.Fields["Genres"].(string),
		}

		if r.highlight && searchReq.Highlight != nil {
			book = r.highlightItem(item.Fragments, book)
		}

		res = append(res, book)
	}

	return res, nil
}
