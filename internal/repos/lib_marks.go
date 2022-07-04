package repos

import (
	"github.com/dgraph-io/badger/v3"
)

type LibMarks struct {
	db *badger.DB
}

func NewLibMarks(db *badger.DB) *LibMarks {
	return &LibMarks{
		db: db,
	}
}

func (r *LibMarks) MarkExists(mark string) (res bool) {
	if err := r.db.View(func(tx *badger.Txn) (txErr error) {
		item, err := tx.Get([]byte(mark))
		if err != nil {
			return err
		}

		item.Value(func(val []byte) error {
			res = string(val) == "true"
			return nil
		})

		return nil
	}); err != nil {
		res = false
	}

	return
}

func (r *LibMarks) AddMark(mark string) error {
	return r.db.Update(func(tx *badger.Txn) (txErr error) {
		return tx.Set([]byte(mark), []byte("true"))
	})
}

func (r *LibMarks) Close() error {
	return r.db.Close()
}
