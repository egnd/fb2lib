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
	"github.com/dgraph-io/badger/v3"
	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/fb2lib/pkg/pagination"
	"github.com/rs/zerolog"
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

type BooksBadgerBleve struct {
	batching bool
	buckets  map[BucketType]*badger.DB
	index    bleve.Index
	encode   entities.IMarshal
	decode   entities.IUnmarshal
	logger   zerolog.Logger
	// cache    *cache.Cache @TODO:
	wg        sync.WaitGroup
	batchPipe chan *entities.Book
	batchStop chan struct{}
}

func NewBooksBadgerBleve(batchSize int,
	buckets map[BucketType]*badger.DB,
	index bleve.Index,
	encode entities.IMarshal,
	decode entities.IUnmarshal,
	logger zerolog.Logger,
) *BooksBadgerBleve {
	repo := &BooksBadgerBleve{
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

func (r *BooksBadgerBleve) getBooks(booksIDs []string) ([]entities.Book, error) {
	res := make([]entities.Book, 0, len(booksIDs))
	err := r.buckets[BucketBooks].View(func(tx *badger.Txn) error {
		var book entities.Book
		for _, itemID := range booksIDs {
			item, err := tx.Get([]byte(itemID))
			if err != nil {
				return err
			}

			if err := item.Value(func(val []byte) error {
				return r.decode(val, &book)
			}); err != nil {
				return err
			}

			res = append(res, book)
		}

		return nil
	})

	return res, err
}

func (r *BooksBadgerBleve) GetByID(bookID string) (*entities.Book, error) {
	if bookID == "" {
		return nil, errors.New("empty book id")
	}

	res, err := r.getBooks([]string{bookID})
	if err != nil {
		return nil, err
	}

	return &res[0], nil
}

func (r *BooksBadgerBleve) FindBooks(queryStr string,
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

func (r *BooksBadgerBleve) Remove(bookID string) error { //@TODO: remove book file too
	if err := r.index.Delete(bookID); err != nil {
		return err
	}

	return r.buckets[BucketBooks].Update(func(tx *badger.Txn) error {
		return tx.Delete([]byte(bookID))
	})
}

func (r *BooksBadgerBleve) clearSeqs(vals []string) []string {
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

func (r *BooksBadgerBleve) buildOrCond(field entities.IndexField, vals []string) query.Query {
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

func (r *BooksBadgerBleve) GetSeriesBooks(limit int, series []string, except *entities.Book) (res []entities.Book, err error) {
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

func (r *BooksBadgerBleve) GetAuthorsBooks(limit int, authors []string, except *entities.Book) (res []entities.Book, err error) {
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

func (r *BooksBadgerBleve) GetAuthorsSeries(authors []string, except []string) (entities.FreqsItems, error) {
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

func (r *BooksBadgerBleve) GetCnt(bucket BucketType) (uint64, error) {
	var res uint64
	err := r.buckets[bucket].View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			res++
		}

		return nil
	})

	return res, err
}

func (r *BooksBadgerBleve) getFreqs(bucket BucketType, prefixes ...string) (entities.FreqsItems, error) {
	res := make(entities.FreqsItems, 0, 500)
	err := r.buckets[bucket].View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		if len(prefixes) == 0 {
			for it.Rewind(); it.Valid(); it.Next() {
				if err := it.Item().Value(func(val []byte) error {
					var freqItem entities.ItemFreq
					if err := r.decode(val, &freqItem); err != nil {
						return nil
					}
					res = append(res, freqItem)
					return nil
				}); err != nil {
					return err
				}
			}
		} else {
			for _, prefix := range prefixes {
				pref := []byte(prefix)
				for it.Seek(pref); it.ValidForPrefix(pref); it.Next() {
					if err := it.Item().Value(func(val []byte) error {
						var freqItem entities.ItemFreq
						if err := r.decode(val, &freqItem); err != nil {
							return nil
						}
						res = append(res, freqItem)
						return nil
					}); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	return res, err
}

func (r *BooksBadgerBleve) GetGenres(pager pagination.IPager) (entities.FreqsItems, error) {
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

func (r *BooksBadgerBleve) GetLibs() (entities.FreqsItems, error) {
	res, err := r.getFreqs(BucketLibs) // @TODO: cache res slice
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	return res, nil
}

func (r *BooksBadgerBleve) GetLangs() (entities.FreqsItems, error) {
	res, err := r.getFreqs(BucketLangs) // @TODO: cache res slice
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	return res, nil
}

func (r *BooksBadgerBleve) GetSeriesByPrefix(prefix string, pager pagination.IPager) (entities.FreqsItems, error) {
	if prefix == "" {
		return nil, nil
	}

	res, err := r.getFreqs( // @TODO: cache res slice
		BucketSeries, strings.ToLower(prefix), strings.ToUpper(prefix),
	)
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

func (r *BooksBadgerBleve) GetAuthorsByPrefix(prefix string, pager pagination.IPager) (entities.FreqsItems, error) {
	if prefix == "" {
		return nil, nil
	}

	res, err := r.getFreqs( // @TODO: cache res slice
		BucketAuthors, strings.ToLower(prefix), strings.ToUpper(prefix),
	)
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

func (r *BooksBadgerBleve) SaveBook(book *entities.Book) (err error) {
	defer func() {
		if r := recover(); err == nil && r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	r.batchPipe <- book

	return
}

func (r *BooksBadgerBleve) Close() error {
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

func (r *BooksBadgerBleve) runBatching(batchSize int) {
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

func (r *BooksBadgerBleve) saveBatch(batch []*entities.Book, indexBatch *bleve.Batch) {
	if len(batch) == 0 {
		return
	}

	logger := r.logger.With().Int("len", len(batch)).Logger()

	if err := r.buckets[BucketBooks].Update(func(txn *badger.Txn) (err error) {
		var itemData []byte
		for _, item := range batch {
			logger := logger.With().Str("lib", item.Lib).Str("item", item.Src).Logger()

			if itemData, err = r.encode(item); err != nil {
				logger.Error().Err(err).Msg("batch err: encode item")
				continue
			}

			if err = txn.Set([]byte(item.ID), itemData); err != nil {
				logger.Error().Err(err).Msg("batch err: save item")
				continue
			}

			if err = indexBatch.Index(item.ID, item.Index()); err != nil {
				logger.Error().Err(err).Msg("batch err: index item")
			}
		}

		return nil
	}); err != nil {
		logger.Error().Err(err).Msg("batch err: save batch")
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

func (r *BooksBadgerBleve) GetTotal() uint64 {
	total, err := r.index.DocCount()
	if err != nil {
		panic(err)
	}

	return total
}

func (r *BooksBadgerBleve) IterateOver(handlers ...func(*entities.Book) error) error {
	return r.buckets[BucketBooks].View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			it.Item().Value(func(val []byte) error {
				var book entities.Book

				if err := r.decode(val, &book); err != nil {
					r.logger.Warn().Err(err).Str("id", string(it.Item().Key())).Msg("decode book")
					return nil
				}

				for _, handler := range handlers {
					if err := handler(&book); err != nil {
						r.logger.Warn().Err(err).Str("id", string(it.Item().Key())).Msg("handle book")
					}
				}

				return nil
			})
		}

		return nil
	})
}

func (r *BooksBadgerBleve) AppendFreqs(bucket BucketType, items entities.ItemFreqMap) error {
	return r.buckets[bucket].Update(func(txn *badger.Txn) error {
		for k, item := range items {
			if oldItem, err := txn.Get([]byte(k)); err == nil {
				oldItem.Value(func(val []byte) error {
					var oldItem entities.ItemFreq
					r.decode(val, &oldItem)
					item.Freq += oldItem.Freq
					return nil
				})
			}

			data, err := r.encode(item)
			if err != nil {
				r.logger.Warn().Err(err).Str("k", item.Val).Int("v", item.Freq).Msg("encode freq")
				continue
			}

			if err := txn.Set([]byte(k), data); err != nil {
				r.logger.Warn().Err(err).Str("k", item.Val).Int("v", item.Freq).Msg("append freq")
			}
		}

		return nil
	})
}
