package repos

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/rs/zerolog"
)

var (
	regexpSpaces = regexp.MustCompile(`\s+`)
)

type BooksIndexBleve struct {
	highlight bool
	batching  bool
	index     entities.ISearchIndex
	logger    zerolog.Logger
	wg        sync.WaitGroup
	batchStop chan struct{}
	batchPipe chan entities.BookIndex
}

func NewBooksIndexBleve(batchSize int,
	highlight bool, index entities.ISearchIndex, logger zerolog.Logger,
) *BooksIndexBleve {
	repo := &BooksIndexBleve{
		batching:  batchSize > 0,
		highlight: highlight,
		index:     index,
		logger:    logger,
	}

	if repo.batching {
		repo.wg.Add(1)
		repo.batchStop = make(chan struct{})
		repo.batchPipe = make(chan entities.BookIndex)
		go repo.runBatching(batchSize)
	}

	return repo
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
			LibName:          item.Fields["LibName"].(string),
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

func (r *BooksIndexBleve) SaveBook(book entities.BookIndex) (err error) {
	if r.batching {
		defer func() {
			if recover() != nil {
				err = r.index.Index(book.ID, book)
			}
		}()

		r.batchPipe <- book

		return
	}

	return r.index.Index(book.ID, book)
}

func (r *BooksIndexBleve) Close() error {
	if r.batching {
		r.batchStop <- struct{}{}
		r.wg.Wait()
		close(r.batchStop)
		close(r.batchPipe)
	}

	return nil
}

func (r *BooksIndexBleve) runBatching(batchSize int) {
	batch := r.index.NewBatch()

	defer func() {
		if batch.Size() > 0 {
			if err := r.index.Batch(batch); err != nil {
				r.logger.Error().Err(err).Msg("index repo batch last")
			} else {
				r.logger.Debug().Msg("index repo batch last")
			}
		}

		r.wg.Done()
	}()

	for {
		select {
		case <-r.batchStop:
			return
		case doc := <-r.batchPipe:
			if err := batch.Index(doc.ID, doc); err != nil {
				r.logger.Error().Err(err).Msg("index repo add to batch")
				continue
			}

			if batch.Size() < batchSize {
				continue
			}

			if err := r.index.Batch(batch); err != nil {
				r.logger.Error().Err(err).Msg("index repo batch")
			} else {
				r.logger.Debug().Msg("index repo batch")
			}

			batch.Reset()
		}
	}
}
