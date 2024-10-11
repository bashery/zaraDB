package engine

import (
	"fmt"
	"log"
	"strconv"

	"go.etcd.io/bbolt"
)

type Store struct {
	db     *bbolt.DB
	lastid map[string]int64
}

var db *Store

func NewDB(path string) *Store {
	// Open a bbolt database
	kv, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		fmt.Println("can not create file ?? why ?")
		log.Fatal(err)
	}

	lastIds := make(map[string]int64, 0)

	err = kv.View(func(tx *bbolt.Tx) error {
		// Iterate over all buckets in the root
		return tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {

			var lastKey = []byte("0")

			// Get the last key in the bucket
			err = kv.View(func(tx *bbolt.Tx) error {
				bucket := tx.Bucket(name)
				if bucket != nil {
					lastKey, _ = bucket.Cursor().Last()
				}
				return nil
			})
			if err != nil {
				return err
			}
			id, _ := strconv.Atoi(string(lastKey))

			lastIds[string(name)] = int64(id)
			return nil
		})

	})

	if err != nil {
		log.Fatal(err)
		return nil
	}

	db = &Store{db: kv, lastid: lastIds}
	for k, v := range db.lastid {
		fmt.Println(k, v)
	}
	return db
}

func (db *Store) getLastKey(bucket []byte) int64 {
	var lastKey = []byte("0")

	// Get the last key from bucket
	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucket)
		if bucket == nil {
			return fmt.Errorf("bucket should not be nil")
		}
		lastKey, _ = bucket.Cursor().Last()
		return nil
	})
	if err != nil {
		log.Println("getLastKey :", err)
	}
	id, _ := strconv.Atoi(string(lastKey))
	return int64(id)
}

func (db *Store) Put(backet, key, val string) (err error) {
	err = db.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(backet))
		if err != nil {
			fmt.Println("err in Put", err)
			return err
		}
		bucket.Put([]byte(key), []byte(val))
		return nil
	})
	return err
}

func (db *Store) Close() {
	db.db.Sync()
	db.db.Close()
}
