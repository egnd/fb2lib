package factories

import (
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
)

// https://discuss.dgraph.io/t/db-write-latency-in-badgerdb-high-when-the-size-of-the-data-increases-exponentially/16192
// https://github.com/dgraph-io/badger/issues/1297
func NewBadgerDB(dir, name string) *badger.DB {
	dir = path.Join(dir, name)

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil // @TODO:
	opts.NumVersionsToKeep = 0
	opts.ValueLogFileSize = 1024 * 1024 * 10

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}

	return db
}
