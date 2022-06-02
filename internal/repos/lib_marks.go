package repos

import (
	"go.etcd.io/bbolt"
)

type LibMarks struct {
	bucketName string
	storage    *bbolt.DB
}

func NewLibMarks(bucketName string, storage *bbolt.DB) *LibMarks {
	if err := storage.Update(func(tx *bbolt.Tx) (txErr error) {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	}); err != nil {
		panic(err)
	}

	return &LibMarks{
		bucketName: bucketName,
		storage:    storage,
	}
}

func (r *LibMarks) MarkExists(mark string) (res bool) {
	if err := r.storage.View(func(tx *bbolt.Tx) (txErr error) {
		res = string(tx.Bucket([]byte(r.bucketName)).Get([]byte(mark))) == "1"

		return nil
	}); err != nil {
		res = false
	}

	return
}

func (r *LibMarks) AddMark(mark string) error {
	return r.storage.Update(func(tx *bbolt.Tx) (txErr error) {
		return tx.Bucket([]byte(r.bucketName)).Put([]byte(mark), []byte("1"))
	})
}
