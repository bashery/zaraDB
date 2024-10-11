package engine

import (
	"fmt"

	"github.com/tidwall/gjson"
	"go.etcd.io/bbolt"
)

func getCollections() string {

	result := `["`

	// Use a read-only transaction to list all top-level buckets.
	err := db.db.View(func(tx *bbolt.Tx) error {
		// Iterate over each bucket in the root.
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			result += string(name) + ", "

			fmt.Printf("Bucket: %s\n", name)
			return nil
		})
	})

	if err != nil {
		return err.Error()
	}

	result = `{"collections": ` + result[:len(result)-2] + `"], "size": "123mb"}`

	return result
}

// deletes collection
func deleteCollection(query gjson.Result) string {
	// TODO return number of deleted objects
	return "not implemented yet"
}

// creates new collection
func createCollection(query gjson.Result) string {
	return "not implemented yet"
}

// Rename renames db.
func renameDB(query gjson.Result) error {
	return nil
}

// Remove remove db to .Trash dir
func removeDB(query gjson.Result) (err error) {
	return nil
}

// CreateDB create db. TODO return this directly
func createDB(query gjson.Result) (string, error) {

	return "not yet", nil
}

// DeleteDB deletes db. (free hard drive).
func deleteDB(query gjson.Result) string {
	return " not yet"
}
