package factories

import (
	"os"
	"path"

	"github.com/syndtr/goleveldb/leveldb"
)

// https://github.com/syndtr/goleveldb
func NewLevelDB(dir string, name string) *leveldb.DB {
	dir = path.Join(dir, name)

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		panic(err)
	}

	return db
}
