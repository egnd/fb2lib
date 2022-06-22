package factories

import (
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
)

func NewBadgerDB(dir, name string) *badger.DB {
	dir = path.Join(dir, name)

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		panic(err)
	}

	return db
}
