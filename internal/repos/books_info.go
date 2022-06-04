package repos

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/egnd/go-fb2parse"
	"github.com/patrickmn/go-cache"
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
	cache     *cache.Cache
	repoLib   entities.IBooksLibraryRepo
}

func NewBooksInfo(batchSize int, highlight bool,
	storage *bbolt.DB, index entities.ISearchIndex, logger zerolog.Logger,
	encoder entities.IMarshal, decoder entities.IUnmarshal, cache *cache.Cache,
	repoLib entities.IBooksLibraryRepo,
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
		cache:     cache,
		repoLib:   repoLib,
	}

	if repo.batching {
		repo.wg.Add(1)
		repo.batchStop = make(chan struct{})
		repo.batchPipe = make(chan entities.BookInfo)
		go repo.runBatching(batchSize)
	}

	return repo
}

func (r *BooksInfo) GetItems(q query.Query, pager pagination.IPager,
	sort []search.SearchSort, highlight *bleve.HighlightRequest, fields ...string,
) ([]entities.BookInfo, error) {
	var req *bleve.SearchRequest

	if pager == nil {
		total, err := r.index.DocCount()
		if err != nil {
			return nil, err
		}

		req = bleve.NewSearchRequestOptions(q, int(total), 0, false)
	} else {
		req = bleve.NewSearchRequestOptions(q, pager.GetPageSize(), pager.GetOffset(), false)
	}

	req.Fields = fields
	req.Highlight = highlight

	if len(sort) > 0 {
		req.Sort = append(req.Sort, sort...)
	}

	searchResults, err := r.index.Search(req)
	if err != nil {
		return nil, err
	}

	if pager != nil {
		pager.SetTotal(searchResults.Total)
	}

	res := make([]entities.BookInfo, 0, len(searchResults.Hits))
	var book entities.BookInfo

	for _, item := range searchResults.Hits {
		if err := r.storage.View(func(tx *bbolt.Tx) error {
			return r.decode(tx.Bucket([]byte(fmt.Sprintf("lib_%s", path.Base(item.Index)))).Get([]byte(item.ID)), &book)
		}); err != nil {
			return nil, err
		}

		book.Index = entities.NewBookIndex(item)
		res = append(res, book)
	}

	return res, nil
}

func (r *BooksInfo) FindByID(id string) (entities.BookInfo, error) {
	if id == "" {
		return entities.BookInfo{}, errors.New("empty book id")
	}

	items, err := r.GetItems(bleve.NewDocIDQuery(
		[]string{id}), nil, nil, nil,
		"isbn", "title", "author", "transl", "serie", "date", "genre", "publ", "lang", "lib",
	)
	if err != nil {
		return entities.BookInfo{}, err
	}

	if len(items) == 0 {
		return entities.BookInfo{}, fmt.Errorf("book %s not found", id)
	}

	return items[0], nil
}

func (r *BooksInfo) FindIn(lib, queryStr string, pager pagination.IPager) ([]entities.BookInfo, error) {
	highlight := bleve.NewHighlightWithStyle("html")
	highlight.Fields = []string{"isbn", "title", "author", "transl", "serie", "date", "genre", "publ", "lang"}

	queryStr = strings.TrimSpace(strings.ToLower(queryStr))

	var searchQ query.Query
	switch {
	case queryStr == "" || queryStr == "*":
		searchQ = bleve.NewMatchAllQuery()
	default:
		queries := make([]query.Query, 0, 2)
		if strings.Contains(queryStr, " ") {
			queries = append(queries, bleve.NewMatchPhraseQuery(queryStr)) // phrase match
		} else {
			queries = append(queries, bleve.NewTermQuery(queryStr)) // exact word match
		}
		if strings.Contains(queryStr, "*") {
			queries = append(queries, bleve.NewWildcardQuery(queryStr)) // wildcard search syntax (*)
		} else {
			queries = append(queries, bleve.NewQueryStringQuery(queryStr)) // extended search syntax https://blevesearch.com/docs/Query-String-Query/
		}
		searchQ = bleve.NewDisjunctionQuery(queries...)
	}

	if lib != "" {
		searchQ = bleve.NewConjunctionQuery(searchQ, bleve.NewQueryStringQuery(fmt.Sprintf("+lib:%s", lib)))
	}

	res, err := r.GetItems(searchQ, pager, []search.SearchSort{&search.SortField{
		Field:   "year",
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	}}, highlight, append(highlight.Fields, "lib")...)

	if err != nil || len(res) == 0 {
		return nil, err
	}

	r.upgradeDetails(res)

	return res, nil
}

func (r *BooksInfo) upgradeDetails(books []entities.BookInfo) {
	if r.repoLib == nil {
		return
	}

	var book fb2parse.FB2File
	var err error
	var wg sync.WaitGroup

	for k, info := range books {
		wg.Add(1)
		go func(k int, info entities.BookInfo) {
			defer wg.Done()
			if book, err = r.repoLib.GetFB2(info); err != nil {
				return
			}

			books[k].ReadDetails(&book)

			if len(books[k].Details.Images) > 0 {
				books[k].Details.Images = books[k].Details.Images[0:1]
			}
		}(k, info)
	}

	wg.Wait()
}

func (r *BooksInfo) GetGenresFreq(limit int) (entities.GenresIndex, error) {
	if res, found := r.cache.Get(fmt.Sprintf("genres_%d", limit)); found {
		return res.(entities.GenresIndex), nil
	}

	genresCnt := make(map[string]uint32, 100)

	total, err := r.index.DocCount()
	if err != nil {
		return nil, err
	}

	searchReq := bleve.NewSearchRequestOptions(bleve.NewMatchAllQuery(), int(total), 0, false)
	searchReq.Fields = []string{"genr"}
	res, err := r.index.Search(searchReq)
	if err != nil {
		return nil, err
	}

	for _, item := range res.Hits {
		if _, ok := item.Fields["genr"].(string); !ok {
			continue
		}

		for _, val := range strings.Split(item.Fields["genr"].(string), entities.IndexFieldSep) {
			genresCnt[val]++
		}
	}

	genres := make(entities.GenresIndex, 0, len(genresCnt))
	for genre, cnt := range genresCnt {
		genres = append(genres, entities.GenreFreq{Name: genre, Cnt: cnt})
	}

	if sort.Sort(sort.Reverse(genres)); limit > 0 && len(genres) > limit {
		genres = genres[:limit]
	}

	r.cache.Set(fmt.Sprintf("genres_%d", limit), genres, 0)

	return genres, nil
}

// func (r *BooksInfo) SearchByAuthor(strQuery string, pager pagination.IPager) ([]entities.BookInfo, error) {
// 	strQuery = strings.TrimSpace(regexpSpaces.ReplaceAllString(strings.ToLower(strQuery), " "))

// 	var q query.Query
// 	if strQuery == "" {
// 		q = bleve.NewMatchAllQuery()
// 	} else {
// 		q = bleve.NewQueryStringQuery("+auth:" + strings.ReplaceAll(strQuery, " ", " +auth:"))
// 	}

// 	var searchReq *bleve.SearchRequest
// 	if pager == nil {
// 		total, err := r.index.DocCount()
// 		if err != nil {
// 			return nil, err
// 		}
// 		searchReq = bleve.NewSearchRequestOptions(q, int(total), 0, false)
// 	} else {
// 		searchReq = bleve.NewSearchRequestOptions(q, pager.GetPageSize(), pager.GetOffset(), false)
// 	}

// 	searchReq.Sort = append(searchReq.Sort, &search.SortField{
// 		Field:   "year",
// 		Desc:    true,
// 		Type:    search.SortFieldAsNumber,
// 		Missing: search.SortFieldMissingLast,
// 	})

// 	if r.highlight {
// 		searchReq.Highlight = bleve.NewHighlightWithStyle("html")
// 	}

// 	return r.getBooks(searchReq, pager)
// }

// func (r *BooksInfo) SearchBySequence(strQuery string, pager pagination.IPager) ([]entities.BookInfo, error) {
// 	strQuery = strings.TrimSpace(regexpSpaces.ReplaceAllString(strings.ToLower(strQuery), " "))

// 	var q query.Query
// 	if strQuery == "" {
// 		q = bleve.NewMatchAllQuery()
// 	} else {
// 		q = bleve.NewQueryStringQuery("+seq:" + strings.ReplaceAll(strQuery, " ", " +seq:"))
// 	}

// 	var searchReq *bleve.SearchRequest
// 	if pager == nil {
// 		total, err := r.index.DocCount()
// 		if err != nil {
// 			return nil, err
// 		}
// 		searchReq = bleve.NewSearchRequestOptions(q, int(total), 0, false)
// 	} else {
// 		searchReq = bleve.NewSearchRequestOptions(q, pager.GetPageSize(), pager.GetOffset(), false)
// 	}

// 	searchReq.Sort = append(searchReq.Sort, &search.SortField{
// 		Field:   "year",
// 		Desc:    true,
// 		Type:    search.SortFieldAsNumber,
// 		Missing: search.SortFieldMissingLast,
// 	})

// 	if r.highlight {
// 		searchReq.Highlight = bleve.NewHighlightWithStyle("html")
// 	}

// 	return r.getBooks(searchReq, pager)
// }

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
	defer r.wg.Done()

	batch := make([]entities.BookInfo, 0, batchSize)
	indexBatch := r.index.NewBatch()

loop:
	for {
		select {
		case <-r.batchStop:
			break loop
		case doc := <-r.batchPipe:
			batch = append(batch, doc)

			if len(batch) < cap(batch) {
				continue
			}

			r.saveBatch(batch, indexBatch)
			batch = batch[:0]
			indexBatch.Reset()
		}
	}

	r.saveBatch(batch, indexBatch)
}

func (r *BooksInfo) saveBatch(batch []entities.BookInfo, indexBatch *bleve.Batch) {
	if len(batch) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	logger := r.logger.With().Str("lib", batch[0].LibName).Logger()

	go func() {
		defer wg.Done()

		for _, book := range batch {
			if err := indexBatch.Index(book.Index.ID, book.Index); err != nil {
				logger.Error().Err(err).Str("item", book.Src).Msg("info repo batch index item")
			}
		}

		if err := r.index.Batch(indexBatch); err != nil {
			logger.Error().Err(err).Msg("info repo batch exec")
		}
	}()

	go r.storage.Update(func(tx *bbolt.Tx) (err error) {
		defer wg.Done()

		bucket, err := tx.CreateBucketIfNotExists([]byte(fmt.Sprintf("lib_%s", batch[0].LibName)))
		if err != nil {
			logger.Error().Err(err).Msg("info repo batch get bucket")
			return
		}

		var bookBytes []byte

		for _, book := range batch {
			if bookBytes, err = r.encode(book); err != nil {
				logger.Error().Err(err).Str("item", book.Src).Msg("info repo batch encode info")
				continue
			}

			if err = bucket.Put([]byte(book.Index.ID), bookBytes); err != nil {
				logger.Error().Err(err).Str("item", book.Src).Msg("info repo batch save info")
			}
		}

		return
	})

	wg.Wait()
	logger.Debug().Msg("info repo batch saved")
}
