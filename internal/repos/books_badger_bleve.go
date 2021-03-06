package repos

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/rs/zerolog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type BucketType string

const (
	BucketAuthors BucketType = "authors"
	BucketBooks   BucketType = "books"
	BucketSeries  BucketType = "series"
	BucketGenres  BucketType = "genres"
	BucketLibs    BucketType = "libs"
	BucketLangs   BucketType = "langs"
)

type BooksLevelBleve struct {
	batching bool
	buckets  map[BucketType]*leveldb.DB
	index    bleve.Index
	encode   entities.IMarshal
	decode   entities.IUnmarshal
	logger   zerolog.Logger
	// cache    *cache.Cache @TODO:
	wg        sync.WaitGroup
	batchPipe chan *entities.Book
	batchStop chan struct{}
}

func NewBooksLevelBleve(batchSize int,
	buckets map[BucketType]*leveldb.DB,
	index bleve.Index,
	encode entities.IMarshal,
	decode entities.IUnmarshal,
	logger zerolog.Logger,
) *BooksLevelBleve {
	repo := &BooksLevelBleve{
		batching: batchSize > 0,
		buckets:  buckets,
		index:    index,
		encode:   encode,
		decode:   decode,
		logger:   logger,
	}

	if repo.batching {
		repo.wg.Add(1)
		repo.batchStop = make(chan struct{})
		repo.batchPipe = make(chan *entities.Book)
		go repo.runBatching(batchSize)
	}

	return repo
}

func (r *BooksLevelBleve) getBooks(booksIDs []string) ([]entities.Book, error) {
	res := make([]entities.Book, 0, len(booksIDs))

	var book entities.Book

	for _, itemID := range booksIDs {
		data, err := r.buckets[BucketBooks].Get([]byte(itemID), nil)
		if err != nil {
			return nil, err
		}

		if err := r.decode(data, &book); err != nil {
			return nil, err
		}

		res = append(res, book)
	}

	return res, nil
}

func (r *BooksLevelBleve) GetByID(bookID string) (*entities.Book, error) {
	if bookID == "" {
		return nil, errors.New("empty book id")
	}

	res, err := r.getBooks([]string{bookID})
	if err != nil {
		return nil, err
	}

	return &res[0], nil
}

func (r *BooksLevelBleve) FindBooks(queryStr string,
	idxField entities.IndexField, idxFieldVal string, pager pagination.IPager,
) ([]entities.Book, error) {
	queryStr = strings.TrimSpace(strings.ToLower(queryStr))

	var searchQ query.Query
	var sortField *search.SortField
	switch {
	case idxField != entities.IdxFUndefined && idxFieldVal != "":
		searchQ = bleve.NewQueryStringQuery(
			fmt.Sprintf(`+%s:"%s" %s`, idxField, idxFieldVal, queryStr),
		)
		sortField = &search.SortField{Desc: true,
			Field:   string(entities.IdxFYear),
			Type:    search.SortFieldAsNumber,
			Missing: search.SortFieldMissingLast,
		}
	case queryStr == "" || queryStr == "*":
		searchQ = bleve.NewMatchAllQuery()
		sortField = &search.SortField{Desc: true,
			Field:   string(entities.IdxFYear),
			Type:    search.SortFieldAsNumber,
			Missing: search.SortFieldMissingLast,
		}
	default:
		searchQ = bleve.NewDisjunctionQuery(
			bleve.NewMatchPhraseQuery(queryStr), // phrase match
			// bleve.NewWildcardQuery(queryStr),    // wildcards syntax
			bleve.NewQueryStringQuery(queryStr), // extended search syntax https://blevesearch.com/docs/Query-String-Query/
		)
		sortField = &search.SortField{
			Field:   string(entities.IdxFTitle),
			Type:    search.SortFieldAsString,
			Missing: search.SortFieldMissingLast,
		}
	}

	req := bleve.NewSearchRequestOptions(searchQ, pager.GetPageSize(), pager.GetOffset(), false)
	req.Sort = append(req.Sort, sortField)
	req.Highlight = bleve.NewHighlightWithStyle("html")

	searchResults, err := r.index.Search(req)
	if err != nil {
		return nil, err
	}

	pager.SetTotal(searchResults.Total)

	ids := make([]string, 0, len(searchResults.Hits))
	fragments := make(map[string]map[string]string, len(searchResults.Hits))
	for _, item := range searchResults.Hits {
		fragments[item.ID] = make(map[string]string, len(item.Fragments))
		for k, vals := range item.Fragments {
			fragments[item.ID][k] = vals[0]
		}

		ids = append(ids, item.ID)
	}

	res, err := r.getBooks(ids)
	if err != nil {
		return res, err
	}

	for k := range res {
		res[k].Match = fragments[res[k].ID]
	}

	return res, err
}

func (r *BooksLevelBleve) Remove(bookID string) error { //@TODO: remove book file too
	if err := r.index.Delete(bookID); err != nil {
		return err
	}

	return r.buckets[BucketBooks].Delete([]byte(bookID), nil)
}

func (r *BooksLevelBleve) clearSeqs(vals []string) []string {
	res := make([]string, 0, len(vals))

	for _, item := range vals {
		if strings.ContainsRune(item, '(') && strings.ContainsRune(item, ')') {
			item = strings.Split(item, "(")[0]
		}
		if item = strings.ToLower(strings.TrimSpace(item)); item != "" {
			res = append(res, item)
		}
	}

	return res
}

func (r *BooksLevelBleve) buildOrCond(field entities.IndexField, vals []string) query.Query {
	items := make([]query.Query, 0, len(vals))

	for _, item := range r.clearSeqs(vals) {
		if item == "" {
			continue
		}
		items = append(items, bleve.NewQueryStringQuery(fmt.Sprintf(`+%s:"%s"`, field, item)))
	}

	if len(items) == 0 {
		return nil
	}

	return bleve.NewDisjunctionQuery(items...)
}

func (r *BooksLevelBleve) GetSeriesBooks(limit int, series []string, except *entities.Book) (res []entities.Book, err error) {
	searchQ := r.buildOrCond(entities.IdxFSerie, series)
	if searchQ == nil {
		return
	}

	if except != nil {
		searchQ = bleve.NewConjunctionQuery(searchQ,
			bleve.NewQueryStringQuery(fmt.Sprintf(`-%s:%s`, entities.IdxFTitle, except.Info.Title)),
		)
	}

	req := bleve.NewSearchRequestOptions(searchQ, limit, 0, false)
	req.Sort = append(req.Sort, &search.SortField{
		Field: string(entities.IdxFSerie), Type: search.SortFieldAsString,
	})
	searchResults, err := r.index.Search(req)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(searchResults.Hits))
	for _, item := range searchResults.Hits {
		ids = append(ids, item.ID)
	}

	return r.getBooks(ids)
}

func (r *BooksLevelBleve) GetAuthorsBooks(limit int, authors []string, except *entities.Book) (res []entities.Book, err error) {
	searchQ := r.buildOrCond(entities.IdxFAuthor, authors)
	if searchQ == nil {
		return
	}

	if except != nil {
		var buf bytes.Buffer
		for _, item := range r.clearSeqs(except.Series()) {
			buf.WriteRune('-')
			buf.WriteString(string(entities.IdxFSerie))
			buf.WriteRune(':')
			buf.WriteRune('"')
			buf.WriteString(item)
			buf.WriteRune('"')
		}
		searchQ = bleve.NewConjunctionQuery(searchQ, bleve.NewQueryStringQuery(buf.String()))
	}

	req := bleve.NewSearchRequestOptions(searchQ, limit, 0, false)
	req.Sort = append(req.Sort, &search.SortField{
		Field: string(entities.IdxFAuthor), Type: search.SortFieldAsString,
	})
	searchResults, err := r.index.Search(req)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(searchResults.Hits))
	for _, item := range searchResults.Hits {
		ids = append(ids, item.ID)
	}

	return r.getBooks(ids)
}

func (r *BooksLevelBleve) GetAuthorsSeries(authors []string, except []string) (entities.FreqsItems, error) {
	books, err := r.GetAuthorsBooks(1000, authors, nil)
	if err != nil {
		return nil, err
	}

	index := map[string]int{}
	for _, book := range books {
		for _, serie := range r.clearSeqs(book.Series()) {
			index[serie]++
		}
	}

	// @TODO: cache res map

	for _, item := range except {
		delete(index, strings.ToLower(item))
	}

	res := make(entities.FreqsItems, 0, len(index))
	for k, v := range index {
		res = append(res, entities.ItemFreq{Val: k, Freq: v})
	}

	return res, nil
}

func (r *BooksLevelBleve) GetCnt(bucket BucketType) (res uint64, err error) {
	iter := r.buckets[bucket].NewIterator(nil, nil)

	for iter.Next() {
		res++
	}

	iter.Release()

	return res, iter.Error()
}

func (r *BooksLevelBleve) getFreqs(bucket BucketType, prefix ...string) (entities.FreqsItems, error) {
	res := make(entities.FreqsItems, 0, 500)

	var iter iterator.Iterator
	if len(prefix) == 0 {
		iter = r.buckets[bucket].NewIterator(nil, nil)
	} else {
		iter = r.buckets[bucket].NewIterator(util.BytesPrefix([]byte(prefix[0])), nil)
	}

	defer iter.Release()

	for iter.Next() {
		var freqItem entities.ItemFreq
		if err := r.decode(iter.Value(), &freqItem); err != nil {
			return nil, err
		}

		res = append(res, freqItem)
	}

	return res, iter.Error()
}

func (r *BooksLevelBleve) GetGenres(pager pagination.IPager) (entities.FreqsItems, error) {
	res, err := r.getFreqs(BucketGenres) // @TODO: cache res slice
	if err != nil {
		return nil, err
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

func (r *BooksLevelBleve) GetLibs() (entities.FreqsItems, error) {
	res, err := r.getFreqs(BucketLibs) // @TODO: cache res slice
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	return res, nil
}

func (r *BooksLevelBleve) GetLangs() (entities.FreqsItems, error) {
	res, err := r.getFreqs(BucketLangs) // @TODO: cache res slice
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	return res, nil
}

func (r *BooksLevelBleve) GetSeriesByPrefix(prefix string, pager pagination.IPager) (entities.FreqsItems, error) {
	if prefix == "" {
		return nil, nil
	}

	res, err := r.getFreqs(BucketSeries, strings.ToLower(prefix)) // @TODO: cache res slice
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	if pager == nil {
		return res, nil
	}

	pager.SetTotal(len(res))

	if len(res) < pager.GetOffset()+pager.GetPageSize() {
		return res[pager.GetOffset():], nil
	}

	return res[pager.GetOffset() : pager.GetOffset()+pager.GetPageSize()], nil
}

func (r *BooksLevelBleve) GetAuthorsByPrefix(prefix string, pager pagination.IPager) (entities.FreqsItems, error) {
	if prefix == "" {
		return nil, nil
	}

	res, err := r.getFreqs(BucketAuthors, strings.ToLower(prefix)) // @TODO: cache res slice
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	if pager == nil {
		return res, nil
	}

	pager.SetTotal(len(res))

	if len(res) < pager.GetOffset()+pager.GetPageSize() {
		return res[pager.GetOffset():], nil
	}

	return res[pager.GetOffset() : pager.GetOffset()+pager.GetPageSize()], nil
}

func (r *BooksLevelBleve) SaveBook(book *entities.Book) (err error) {
	defer func() {
		if r := recover(); err == nil && r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	r.batchPipe <- book

	return
}

func (r *BooksLevelBleve) Close() error {
	if r.batching {
		r.batchStop <- struct{}{}
		r.wg.Wait()
		close(r.batchStop)
		close(r.batchPipe)
	}

	for _, bucketName := range []BucketType{BucketBooks, BucketAuthors, BucketSeries, BucketGenres, BucketLibs} {
		if bucket, ok := r.buckets[bucketName]; ok && bucket != nil {
			if err := bucket.Close(); err != nil {
				r.logger.Error().Err(err).Str("bucket", string(bucketName)).Msg("close bucket")
			}
		}
	}

	if err := r.index.Close(); err != nil {
		return err
	}

	return nil
}

func (r *BooksLevelBleve) runBatching(batchSize int) {
	defer r.wg.Done()

	batch := make([]*entities.Book, 0, batchSize)
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

func (r *BooksLevelBleve) saveBatch(batch []*entities.Book, indexBatch *bleve.Batch) {
	if len(batch) == 0 {
		return
	}

	logger := r.logger.With().Int("len", len(batch)).Logger()

	var itemData []byte
	var err error

	for _, item := range batch {
		logger := logger.With().Str("lib", item.Lib).Str("item", item.Src).Logger()

		if itemData, err = r.encode(item); err != nil {
			logger.Error().Err(err).Msg("batch err: encode item")
			continue
		}

		if err = r.buckets[BucketBooks].Put([]byte(item.ID), itemData, nil); err != nil {
			logger.Error().Err(err).Msg("batch err: save item")
			continue
		}

		if err = indexBatch.Index(item.ID, item.Index()); err != nil {
			logger.Error().Err(err).Msg("batch err: index item")
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if indexBatch.Size() > 0 {
			if err := r.index.Batch(indexBatch); err != nil {
				logger.Error().Err(err).Msg("batch err: index batch")
			}
		}
	}()

	wg.Wait()
	logger.Debug().Msg("batch saved")
}

func (r *BooksLevelBleve) GetTotal() (total uint64) {
	iter := r.buckets[BucketBooks].NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		total++
	}

	if err := iter.Error(); err != nil {
		panic(err)
	}

	return
}

func (r *BooksLevelBleve) IterateOver(handlers ...func(*entities.Book) error) error {
	iter := r.buckets[BucketBooks].NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		var book entities.Book

		if err := r.decode(iter.Value(), &book); err != nil {
			r.logger.Warn().Err(err).Str("id", string(iter.Key())).Msg("decode book")
			return nil
		}

		for _, handler := range handlers {
			if err := handler(&book); err != nil {
				r.logger.Warn().Err(err).Str("id", string(iter.Key())).Msg("handle book")
			}
		}
	}

	return iter.Error()
}

func (r *BooksLevelBleve) AppendFreqs(bucket BucketType, items entities.ItemFreqMap) (err error) {
	var data []byte
	for k, item := range items {
		if data, err = r.buckets[bucket].Get([]byte(k), nil); err == nil {
			var oldItem entities.ItemFreq
			r.decode(data, &oldItem)
			item.Freq += oldItem.Freq
		}

		if data, err = r.encode(item); err != nil {
			r.logger.Warn().Err(err).Str("k", item.Val).Int("v", item.Freq).Msg("encode freq")
			continue
		}

		if err = r.buckets[bucket].Put([]byte(k), data, nil); err != nil {
			r.logger.Warn().Err(err).Str("k", item.Val).Int("v", item.Freq).Msg("append freq")
		}
	}

	return nil
}
