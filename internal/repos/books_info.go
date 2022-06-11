package repos

import (
	"bytes"
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
	"github.com/egnd/go-xmlparse/fb2"
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
	libs      entities.Libraries
}

func NewBooksInfo(batchSize int, highlight bool,
	storage *bbolt.DB, index entities.ISearchIndex, logger zerolog.Logger,
	encoder entities.IMarshal, decoder entities.IUnmarshal, cache *cache.Cache,
	repoLib entities.IBooksLibraryRepo, libs entities.Libraries,
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
		libs:      libs,
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
		entities.IdxFieldISBN, entities.IdxFieldTitle, entities.IdxFieldAuthor, entities.IdxFieldTranslator, entities.IdxFieldSerie,
		entities.IdxFieldDate, entities.IdxFieldGenre, entities.IdxFieldPublisher, entities.IdxFieldLang, entities.IdxFieldLib,
	)
	if err != nil {
		return entities.BookInfo{}, err
	}

	if len(items) == 0 {
		return entities.BookInfo{}, fmt.Errorf("book %s not found", id)
	}

	items = items[:1]
	r.upgradeDetails(items)

	return items[0], nil
}

func (r *BooksInfo) FindBooks(queryStr, tagName, tagValue string, pager pagination.IPager) ([]entities.BookInfo, error) {
	highlight := bleve.NewHighlightWithStyle("html")
	queryStr = strings.TrimSpace(strings.ToLower(queryStr))
	sortFields := []search.SearchSort{&search.SortField{ // @TODO: sort by title
		Field:   entities.IdxFieldYear,
		Desc:    true,
		Type:    search.SortFieldAsNumber,
		Missing: search.SortFieldMissingLast,
	}}

	var searchQ query.Query
	switch {
	case queryStr == "" || queryStr == "*":
		searchQ = bleve.NewMatchAllQuery()
	default:
		sortFields = nil

		searchQ = bleve.NewDisjunctionQuery(
			bleve.NewMatchPhraseQuery(queryStr), // phrase match
			bleve.NewQueryStringQuery(queryStr), // extended search syntax https://blevesearch.com/docs/Query-String-Query/
		)
	}

	fields := []string{entities.IdxFieldISBN, entities.IdxFieldTitle, entities.IdxFieldAuthor,
		entities.IdxFieldTranslator, entities.IdxFieldSerie, entities.IdxFieldDate,
		entities.IdxFieldGenre, entities.IdxFieldPublisher, entities.IdxFieldLang, entities.IdxFieldLib}

	if tagName != "" && tagValue != "" {
		searchQ = bleve.NewConjunctionQuery(searchQ, bleve.NewQueryStringQuery(fmt.Sprintf(`+%s:"%s"`, tagName, tagValue)))
		for _, item := range fields {
			if item == tagName {
				continue
			}
			highlight.Fields = append(highlight.Fields, item)
		}
	}

	res, err := r.GetItems(searchQ, pager, sortFields, highlight, fields...)

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

	var book fb2.File
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

func (r *BooksInfo) Remove(bookID string) error { //@TODO: remove book file too
	if err := r.index.Delete(bookID); err != nil {
		return err
	}

	return r.storage.Update(func(tx *bbolt.Tx) error {
		for lib := range r.libs {
			bucket := tx.Bucket([]byte(lib))
			if bucket == nil {
				continue
			}

			if err := bucket.Delete([]byte(bookID)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *BooksInfo) GetSeriesBooks(series string, curBook *entities.BookInfo) (res []entities.BookInfo, err error) {
	if series == "" {
		return
	}

	var searchQ query.Query
	var conditions []query.Query
	for _, item := range strings.Split(series, entities.IndexFieldSep) {
		conditions = append(conditions, bleve.NewQueryStringQuery(
			fmt.Sprintf(`+serie:"%s"`, strings.TrimSpace(strings.Split(item, "(")[0])),
		))
	}

	searchQ = bleve.NewDisjunctionQuery(conditions...)

	if curBook != nil {
		searchQ = bleve.NewConjunctionQuery(searchQ, bleve.NewQueryStringQuery(fmt.Sprintf(`-ID:%s`, curBook.Index.ID)))
	}

	if res, err = r.GetItems(searchQ, nil, []search.SearchSort{&search.SortField{
		Field: entities.IdxFieldSerie, Type: search.SortFieldAsString,
	}}, nil, entities.IdxFieldISBN, entities.IdxFieldTitle, entities.IdxFieldAuthor, entities.IdxFieldTranslator,
		entities.IdxFieldSerie, entities.IdxFieldDate, entities.IdxFieldGenre, entities.IdxFieldPublisher,
		entities.IdxFieldLang, entities.IdxFieldLib); err != nil {
		return nil, err
	}

	r.upgradeDetails(res)

	return res, nil
}

func (r *BooksInfo) GetOtherAuthorBooks(authors string, curBook *entities.BookInfo) (res []entities.BookInfo, err error) {
	if authors == "" {
		return
	}

	var searchQ query.Query
	var conditions []query.Query
	for _, item := range strings.Split(authors, entities.IndexFieldSep) {
		conditions = append(conditions, bleve.NewQueryStringQuery(
			fmt.Sprintf(`+author:"%s"`, strings.TrimSpace(item)),
		))
	}

	searchQ = bleve.NewDisjunctionQuery(conditions...)

	if curBook != nil {
		var buf bytes.Buffer
		for _, item := range strings.Split(curBook.Index.Serie, entities.IndexFieldSep) {
			buf.WriteString(`-serie:"`)
			buf.WriteString(strings.TrimSpace(strings.Split(item, "(")[0]))
			buf.WriteString(`" `)
		}
		searchQ = bleve.NewConjunctionQuery(searchQ, bleve.NewQueryStringQuery(buf.String()))
	}

	if res, err = r.GetItems(searchQ, nil, nil, nil, entities.IdxFieldISBN, entities.IdxFieldTitle,
		entities.IdxFieldAuthor, entities.IdxFieldTranslator, entities.IdxFieldSerie, entities.IdxFieldDate,
		entities.IdxFieldGenre, entities.IdxFieldPublisher, entities.IdxFieldLang, entities.IdxFieldLib,
	); err != nil {
		return nil, err
	}

	r.upgradeDetails(res)

	return res, nil
}

func (r *BooksInfo) GetOtherAuthorSeries(authors, curSeries string) (res map[string]int, err error) {
	if authors == "" {
		return
	}

	var searchQ query.Query
	var conditions []query.Query
	for _, item := range strings.Split(authors, entities.IndexFieldSep) {
		conditions = append(conditions, bleve.NewQueryStringQuery(
			fmt.Sprintf(`+author:"%s"`, strings.TrimSpace(item)),
		))
	}
	searchQ = bleve.NewDisjunctionQuery(conditions...)

	var books []entities.BookInfo
	if books, err = r.GetItems(searchQ, nil, nil, nil, entities.IdxFieldSerie); err != nil {
		return
	}

	res = map[string]int{}
	for _, item := range books {
		for _, serie := range strings.Split(item.Index.Serie, entities.IndexFieldSep) {
			serie = strings.Split(serie, "(")[0]
			if strings.Contains(curSeries, serie) {
				continue
			}

			res[serie]++
		}
	}

	return res, nil
}

func (r *BooksInfo) GetStats() (map[string]uint64, error) {
	if cachedRes, found := r.cache.Get("index_stats"); found {
		return cachedRes.(map[string]uint64), nil
	}

	var res map[string]uint64

	defer func() {
		r.cache.Add("index_stats", res, 0)
	}()

	total, err := r.index.DocCount()
	if err != nil {
		return nil, err
	}

	searchReq := bleve.NewSearchRequestOptions(bleve.NewMatchAllQuery(), int(total), 0, false)
	searchReq.Fields = []string{entities.IdxFieldGenre, entities.IdxFieldAuthor, entities.IdxFieldSerie}
	items, err := r.index.Search(searchReq)
	if err != nil {
		return nil, err
	}

	genres := map[string]uint32{}
	authors := map[string]uint32{}
	series := map[string]uint32{}

	for _, item := range items.Hits {
		if genre, ok := item.Fields[entities.IdxFieldGenre].(string); ok {
			for _, val := range strings.Split(genre, entities.IndexFieldSep) {
				genres[val]++
			}
		}

		if author, ok := item.Fields[entities.IdxFieldAuthor].(string); ok {
			for _, val := range strings.Split(author, entities.IndexFieldSep) {
				authors[val]++
			}
		}

		if serie, ok := item.Fields[entities.IdxFieldSerie].(string); ok {
			for _, val := range strings.Split(serie, entities.IndexFieldSep) {
				series[strings.Split(val, " (")[0]]++
			}
		}
	}

	res = map[string]uint64{
		"books":   total,
		"genres":  uint64(len(genres)),
		"authors": uint64(len(authors)),
		"series":  uint64(len(series)),
	}

	return res, err
}

func (r *BooksInfo) GetGenres(pager pagination.IPager) (entities.GenresIndex, error) {
	if cachedRes, found := r.cache.Get("genres"); found {
		if pager == nil {
			return cachedRes.(entities.GenresIndex), nil
		}

		pager.SetTotal(len(cachedRes.(entities.GenresIndex)))

		if int(pager.GetTotal()) < pager.GetOffset()+pager.GetPageSize() {
			return cachedRes.(entities.GenresIndex)[pager.GetOffset():], nil
		}

		return cachedRes.(entities.GenresIndex)[pager.GetOffset() : pager.GetOffset()+pager.GetPageSize()], nil
	}

	var res entities.GenresIndex

	defer func() {
		r.cache.Set("genres", res, 0)
	}()

	total, err := r.index.DocCount()
	if err != nil {
		return nil, err
	}

	searchReq := bleve.NewSearchRequestOptions(bleve.NewMatchAllQuery(), int(total), 0, false)
	searchReq.Fields = []string{entities.IdxFieldGenre}
	items, err := r.index.Search(searchReq)
	if err != nil {
		return nil, err
	}

	genresFreq := make(map[string]uint32, 300)
	for _, item := range items.Hits {
		if genre, ok := item.Fields[entities.IdxFieldGenre].(string); ok {
			for _, val := range strings.Split(genre, entities.IndexFieldSep) {
				genresFreq[val]++
			}
		}
	}

	for genre, cnt := range genresFreq {
		res = append(res, entities.GenreFreq{Name: genre, Cnt: cnt})
	}

	sort.Sort(sort.Reverse(res))

	if pager == nil {
		return res, nil
	}

	pager.SetTotal(len(res))

	if len(res) < pager.GetOffset()+pager.GetPageSize() {
		return res[pager.GetOffset():], nil
	}

	return res[pager.GetOffset() : pager.GetOffset()+pager.GetPageSize()], nil
}

func (r *BooksInfo) GetSeriesByLetter(letter string) ([]string, error) {
	letter = strings.TrimSpace(strings.ToLower(letter))

	if letter == "" {
		return nil, nil
	}

	if cachedRes, found := r.cache.Get(fmt.Sprintf("series_starts_%s", letter)); found {
		return cachedRes.([]string), nil
	}

	items, err := r.GetItems(bleve.NewQueryStringQuery(fmt.Sprintf(`+serie:%s*`, letter)), nil,
		[]search.SearchSort{&search.SortField{Field: entities.IdxFieldSerie, Type: search.SortFieldAsString}},
		nil, entities.IdxFieldSerie)
	if err != nil {
		return nil, err
	}

	dupl := make(map[string]struct{}, len(items))
	res := make([]string, 0, len(items))

	for _, item := range items {
		for _, serie := range strings.Split(item.Index.Serie, entities.IndexFieldSep) {
			serie = strings.Split(serie, " (")[0]
			if _, ok := dupl[serie]; ok || !strings.HasPrefix(strings.ToLower(serie), letter) {
				continue
			}
			dupl[serie] = struct{}{}
			res = append(res, serie)
		}
	}

	r.cache.Set(fmt.Sprintf("series_starts_%s", letter), res, 0)

	return res, nil
}

func (r *BooksInfo) GetAuthorsByLetter(letter string) ([]string, error) {
	letter = strings.TrimSpace(strings.ToLower(letter))

	if letter == "" {
		return nil, nil
	}

	if cachedRes, found := r.cache.Get(fmt.Sprintf("authors_starts_%s", letter)); found {
		return cachedRes.([]string), nil
	}

	items, err := r.GetItems(bleve.NewQueryStringQuery(fmt.Sprintf("+%s:%s*", entities.IdxFieldAuthor, letter)), nil,
		[]search.SearchSort{&search.SortField{Field: entities.IdxFieldAuthor, Type: search.SortFieldAsString}},
		nil, entities.IdxFieldAuthor)
	if err != nil {
		return nil, err
	}

	dupl := make(map[string]struct{}, len(items))
	res := make([]string, 0, len(items))

	for _, item := range items {
		for _, author := range strings.Split(item.Index.Author, entities.IndexFieldSep) {
			if _, ok := dupl[author]; ok || !strings.HasPrefix(strings.ToLower(author), letter) {
				continue
			}
			dupl[author] = struct{}{}
			res = append(res, author)
		}
	}

	r.cache.Set(fmt.Sprintf("authors_starts_%s", letter), res, 0)

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
