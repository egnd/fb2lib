package factories

import (
	"os"
	"path"

	"go.etcd.io/bbolt"
)

func NewBoltDB(dbPath string) *bbolt.DB {
	if err := os.MkdirAll(path.Dir(dbPath), 0755); err != nil {
		panic(err)
	}

	db, err := bbolt.Open(dbPath, 0644, nil)
	if err != nil {
		panic(err)
	}

	return db
}
