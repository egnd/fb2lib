package repos

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/rs/zerolog"
	"go.etcd.io/bbolt"
)

var (
	regexpSpaces = regexp.MustCompile(`\s+`)
)

type BooksInfo struct {
	highlight bool
	batching  bool
	index     entities.ISearchIndex
	storage   *bbolt.DB
	logger    zerolog.Logger
	wg        sync.WaitGroup
	batchStop chan struct{}
	batchPipe chan entities.BookInfo
	encode    entities.IMarshal
	decode    entities.IUnmarshal
}

func NewBooksInfo(batchSize int, highlight bool,
	storage *bbolt.DB, index entities.ISearchIndex, logger zerolog.Logger,
	encoder entities.IMarshal, decoder entities.IUnmarshal,
) *BooksInfo {
	if batchSize > storage.MaxBatchSize {
		storage.MaxBatchSize = batchSize
	}

	repo := &BooksInfo{
		batching:  batchSize > 0,
		highlight: highlight,
		index:     index,
		logger:    logger,
		storage:   storage,
		encode:    encoder,
		decode:    decoder,
	}

	if repo.batching {
		repo.wg.Add(1)
		repo.batchStop = make(chan struct{})
		repo.batchPipe = make(chan entities.BookInfo)
		go repo.runBatching(batchSize)
	}

	return repo
}

func (r *BooksInfo) SearchAll(strQuery string, pager pagination.IPager) ([]entities.BookInfo, error) {
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
		Field:   "year",
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	})

	return r.getBooks(searchReq, pager)
}

func (r *BooksInfo) SearchByAuthor(strQuery string, pager pagination.IPager) ([]entities.BookInfo, error) {
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
		Field:   "year",
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	})

	return r.getBooks(searchReq, pager)
}

func (r *BooksInfo) SearchBySequence(strQuery string, pager pagination.IPager) ([]entities.BookInfo, error) {
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
		Field:   "year",
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	})

	return r.getBooks(searchReq, pager)
}

func (r *BooksInfo) GetBook(bookID string) (entities.BookInfo, error) {
	if bookID == "" {
		return entities.BookInfo{}, errors.New("repo get book error: empty book id")
	}

	searchReq := bleve.NewSearchRequestOptions(
		bleve.NewDocIDQuery([]string{bookID}), 1, 0, false,
	)

	items, err := r.getBooks(searchReq, nil)
	if err != nil {
		return entities.BookInfo{}, err
	}

	if len(items) == 0 {
		return entities.BookInfo{}, fmt.Errorf("repo getbook error: %s not found", bookID)
	}

	return items[0], nil
}

func (r *BooksInfo) highlightItem(fragments search.FieldFragmentMap, book entities.BookIndex) entities.BookIndex {
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

func (r *BooksInfo) getBooks(
	searchReq *bleve.SearchRequest, pager pagination.IPager,
) ([]entities.BookInfo, error) {
	searchReq.Fields = []string{"*"}

	searchResults, err := r.index.Search(searchReq)
	if err != nil {
		return nil, err
	}

	if pager != nil {
		pager.SetTotal(searchResults.Total)
	}

	res := make([]entities.BookInfo, 0, len(searchResults.Hits))

	for _, item := range searchResults.Hits {
		var book entities.BookInfo
		if err := r.storage.View(func(tx *bbolt.Tx) error {
			return r.decode(tx.Bucket([]byte(fmt.Sprintf("lib_%s", path.Base(item.Index)))).Get([]byte(item.ID)), &book)
		}); err != nil {
			continue
		}

		book.Index = entities.BookIndex{
			ID:        item.ID,
			Lang:      item.Fields["lng"].(string),
			ISBN:      item.Fields["isbn"].(string),
			Titles:    item.Fields["name"].(string),
			Authors:   item.Fields["auth"].(string),
			Sequences: item.Fields["seq"].(string),
			Publisher: item.Fields["publ"].(string),
			Date:      item.Fields["date"].(string),
			Genres:    item.Fields["genr"].(string),
		}

		if r.highlight && searchReq.Highlight != nil {
			book.Index = r.highlightItem(item.Fragments, book.Index)
		}

		res = append(res, book)
	}

	return res, nil
}

func (r *BooksInfo) SaveBook(book entities.BookInfo) (err error) {
	if r.batching {
		defer func() {
			if r := recover(); err == nil && r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()

		r.batchPipe <- book

		return
	}

	if err = r.index.Index(book.Index.ID, book.Index); err == nil {
		err = r.storage.Update(func(tx *bbolt.Tx) (txErr error) {
			bookBytes, txErr := r.encode(book)
			if txErr != nil {
				return fmt.Errorf("encode book error: %s", txErr)
			}

			b, txErr := tx.CreateBucketIfNotExists([]byte(fmt.Sprintf("lib_%s", book.LibName)))
			if txErr != nil {
				return fmt.Errorf("create bucket: %s", txErr)
			}

			return b.Put([]byte(book.Index.ID), bookBytes)
		})
	}

	return
}

func (r *BooksInfo) Close() error {
	if r.batching {
		r.batchStop <- struct{}{}
		r.wg.Wait()
		close(r.batchStop)
		close(r.batchPipe)
	}

	return r.index.Close()
}

func (r *BooksInfo) runBatching(batchSize int) {
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
			if err := batch.Index(doc.Index.ID, doc.Index); err != nil {
				r.logger.Error().Err(err).Msg("index repo add to batch")
				continue
			}

			if err := r.storage.Batch(func(tx *bbolt.Tx) (txErr error) { // @TODO: improve batch fn
				bookBytes, txErr := r.encode(doc)
				if txErr != nil {
					return fmt.Errorf("encode book error: %s", txErr)
				}

				b, txErr := tx.CreateBucketIfNotExists([]byte(fmt.Sprintf("lib_%s", doc.LibName)))
				if txErr != nil {
					return fmt.Errorf("create bucket: %s", txErr)
				}

				return b.Put([]byte(doc.Index.ID), bookBytes)
			}); err != nil {
				r.logger.Error().Err(err).Msg("book to storage batch")
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
