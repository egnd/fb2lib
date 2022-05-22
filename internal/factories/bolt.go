package factories

import (
	"go.etcd.io/bbolt"
)

func NewBoltDB(dir string) *bbolt.DB {
	db, err := bbolt.Open(dir, 0644, nil)
	if err != nil {
		panic(err)
	}

	return db
}
