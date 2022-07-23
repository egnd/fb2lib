package repos

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type LibMarks struct {
	db *leveldb.DB
}

func NewLibMarks(db *leveldb.DB) *LibMarks {
	return &LibMarks{
		db: db,
	}
}

func (r *LibMarks) MarkExists(mark string) bool {
	data, err := r.db.Get([]byte(mark), nil)
	if err != nil {
		return false
	}

	return string(data) == "true"
}

func (r *LibMarks) AddMark(mark string) error {
	return r.db.Put([]byte(mark), []byte("true"), nil)
}

func (r *LibMarks) Close() error {
	return r.db.Close()
}
