package factories

import (
	"os"
	"path"

	"go.etcd.io/bbolt"
)

func NewBoltDB(dir, name string) *bbolt.DB {
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	db, err := bbolt.Open(path.Join(dir, name), 0644, nil)
	if err != nil {
		panic(err)
	}

	return db
}
