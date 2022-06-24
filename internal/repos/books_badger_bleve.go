package repos

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
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

type BooksBadgerBleve struct {
	batching  bool
	dbBooks   *badger.DB
	dbAuthors *badger.DB
	dbSeries  *badger.DB
	dbGenres  *badger.DB
	dbLibs    *badger.DB
	index     bleve.Index
	encode    entities.IMarshal
	decode    entities.IUnmarshal
	logger    zerolog.Logger
	// cache    *cache.Cache @TODO:
	wg        sync.WaitGroup
	batchPipe chan *entities.Book
	batchStop chan struct{}
}

func NewBooksBadgerBleve(batchSize int,
	dbBooks *badger.DB,
	dbAuthors *badger.DB,
	dbSeries *badger.DB,
	dbGenres *badger.DB,
	dbLibs *badger.DB,
	index bleve.Index,
	encode entities.IMarshal,
	decode entities.IUnmarshal,
	logger zerolog.Logger,
) *BooksBadgerBleve {
	repo := &BooksBadgerBleve{
		batching:  batchSize > 0,
		dbBooks:   dbBooks,
		dbAuthors: dbAuthors,
		dbSeries:  dbSeries,
		dbGenres:  dbGenres,
		dbLibs:    dbLibs,
		index:     index,
		encode:    encode,
		decode:    decode,
		logger:    logger,
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
	err := r.dbBooks.View(func(tx *badger.Txn) error {
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
	case queryStr == "" || queryStr == "*":
		searchQ = bleve.NewMatchAllQuery()
		sortField = &search.SortField{
			Desc:    true,
			Field:   string(entities.IdxFYear),
			Type:    search.SortFieldAsNumber,
			Missing: search.SortFieldMissingLast,
		}
	default:
		searchQ = bleve.NewDisjunctionQuery(
			bleve.NewMatchPhraseQuery(queryStr), // phrase match
			bleve.NewQueryStringQuery(queryStr), // extended search syntax https://blevesearch.com/docs/Query-String-Query/
		)
		sortField = &search.SortField{
			Field:   string(entities.IdxFTitle),
			Type:    search.SortFieldAsString,
			Missing: search.SortFieldMissingLast,
		}
	}

	if idxField != entities.IdxFUndefined && idxFieldVal != "" {
		searchQ = bleve.NewConjunctionQuery(searchQ,
			bleve.NewQueryStringQuery(fmt.Sprintf(`+%s:"%s"`, idxField, idxFieldVal)),
		)
	}

	req := bleve.NewSearchRequestOptions(searchQ, pager.GetPageSize(), pager.GetOffset(), false)
	req.Sort = append(req.Sort, sortField)

	searchResults, err := r.index.Search(req)
	if err != nil {
		return nil, err
	}

	pager.SetTotal(searchResults.Total)

	ids := make([]string, 0, len(searchResults.Hits))
	for _, item := range searchResults.Hits {
		ids = append(ids, item.ID)
	}

	return r.getBooks(ids)
}

func (r *BooksBadgerBleve) Remove(bookID string) error { //@TODO: remove book file too
	if err := r.index.Delete(bookID); err != nil {
		return err
	}

	return r.dbBooks.Update(func(tx *badger.Txn) error {
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
			bleve.NewQueryStringQuery(fmt.Sprintf(`-%s:%s`, entities.IdxFID, except.ID)),
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

func (r *BooksBadgerBleve) GetAuthorsSeries(authors []string, except []string) (res map[string]int, err error) {
	books, err := r.GetAuthorsBooks(1000, authors, nil)
	if err != nil {
		return nil, err
	}

	res = map[string]int{}
	for _, book := range books {
		for _, serie := range r.clearSeqs(book.Series()) {
			res[serie]++
		}
	}

	// @TODO: cache res map

	for _, item := range except {
		delete(res, item)
	}

	return res, nil
}

func (r *BooksBadgerBleve) getCnt(db *badger.DB) (uint64, error) {
	var res uint64
	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			res++
		}

		return nil
	})

	return res, err
}

func (r *BooksBadgerBleve) GetBooksCnt() (uint64, error) {
	return r.getCnt(r.dbBooks)
}

func (r *BooksBadgerBleve) GetAuthorsCnt() (uint64, error) {
	return r.getCnt(r.dbAuthors)
}

func (r *BooksBadgerBleve) GetGenresCnt() (uint64, error) {
	return r.getCnt(r.dbGenres)
}

func (r *BooksBadgerBleve) GetSeriesCnt() (uint64, error) {
	return r.getCnt(r.dbSeries)
}

func (r *BooksBadgerBleve) getFreqs(db *badger.DB, prefixes ...string) (entities.FreqsItems, error) {
	res := make(entities.FreqsItems, 0, 500)
	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		if len(prefixes) == 0 {
			for it.Rewind(); it.Valid(); it.Next() {
				if err := it.Item().Value(func(val []byte) error {
					freq, err := strconv.Atoi(string(val))
					if err != nil {
						return nil
					}

					res = append(res, entities.ItemFreq{
						Freq: freq,
						Val:  string(it.Item().Key()),
					})

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
						freq, err := strconv.Atoi(string(val))
						if err != nil {
							return nil
						}

						res = append(res, entities.ItemFreq{
							Freq: freq,
							Val:  string(it.Item().Key()),
						})

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
	res, err := r.getFreqs(r.dbGenres) // @TODO: cache res slice
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
	res, err := r.getFreqs(r.dbLibs) // @TODO: cache res slice
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	return res, nil
}

func (r *BooksBadgerBleve) GetSeriesByChar(char rune) (entities.FreqsItems, error) {
	res, err := r.getFreqs( // @TODO: cache res slice
		r.dbSeries, strings.ToLower(string(char)),
	)
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	return res, nil
}

func (r *BooksBadgerBleve) GetAuthorsByChar(char rune) (entities.FreqsItems, error) {
	res, err := r.getFreqs( // @TODO: cache res slice
		r.dbAuthors, strings.ToLower(string(char)),
	)
	if err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Val < res[j].Val })

	return res, nil
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

	if err := r.dbBooks.Close(); err != nil {
		return err
	}

	if err := r.dbAuthors.Close(); err != nil {
		return err
	}

	if err := r.dbSeries.Close(); err != nil {
		return err
	}

	if err := r.dbGenres.Close(); err != nil {
		return err
	}

	if err := r.dbLibs.Close(); err != nil {
		return err
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
	summary := map[*badger.DB]map[string]int{
		r.dbAuthors: make(map[string]int, 500),
		r.dbSeries:  make(map[string]int, 500),
		r.dbGenres:  make(map[string]int, 500),
		r.dbLibs:    make(map[string]int, 10),
	}

	if err := r.dbBooks.Update(func(txn *badger.Txn) (err error) {
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

			for _, str := range r.clearSeqs(item.Authors()) {
				summary[r.dbAuthors][str]++
			}
			for _, str := range r.clearSeqs(item.Series()) {
				summary[r.dbSeries][str]++
			}
			for _, str := range r.clearSeqs(item.Genres()) {
				summary[r.dbGenres][str]++
			}
			if item.Lib != "" {
				summary[r.dbLibs][item.Lib]++
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
	wg.Add(5)

	go func() {
		defer wg.Done()
		if indexBatch.Size() > 0 {
			if err := r.index.Batch(indexBatch); err != nil {
				logger.Error().Err(err).Msg("batch err: index batch")
			}
		}
	}()

	for db, items := range summary {
		go func(db *badger.DB, items map[string]int) {
			defer wg.Done()
			db.Update(func(txn *badger.Txn) error {
				for val, cnt := range items {
					if item, err := txn.Get([]byte(val)); err == nil {
						item.Value(func(val []byte) error {
							oldCnt, _ := strconv.Atoi(string(val))
							cnt += oldCnt
							return nil
						})
					}

					if err := txn.Set([]byte(val), []byte(strconv.Itoa(cnt))); err != nil {
						logger.Error().Err(err).Str("val", val).Msg("batch err: save summary item")
					}
				}
				return nil
			})
		}(db, items)
	}

	wg.Wait()
	logger.Debug().Msg("batch saved")
}
