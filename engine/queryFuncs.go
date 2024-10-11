package engine

import (
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
	"go.etcd.io/bbolt"
)

func (s *Store) getData(query gjson.Result) (data []string, err error) {
	coll := query.Get("collection").Str
	if coll == "" {
		return nil, fmt.Errorf(`{"error":"forgot collection name "}`)
	}

	skip := query.Get("skip").Int()
	limit := query.Get("limit").Int()
	if limit == 0 {
		limit = 1000 // what is default setting ?
	}

	// bbolt
	err = s.db.View(func(tx *bbolt.Tx) error {

		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("collection %s is not exists", coll)
		}
		isMatch := query.Get("match")
		// Use a cursor to iterate over all key-value pairs in the bucket.
		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {

			if limit == 0 {
				break
			}

			ok, err := match(isMatch, string(value))
			if err != nil {
				return err
			}

			if ok {
				if skip != 0 {
					skip--
					continue
				}
				data = append(data, string(value))
				limit--
			}

		}

		return nil
	})

	return data, err
}

// Finds first obj match creteria.
func (s *Store) findOne(query gjson.Result) (res string) {
	coll := query.Get("collection").Str
	skip := query.Get("skip").Int()

	err := s.db.View(func(tx *bbolt.Tx) error {

		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", "myBucket")
		}
		isMatch := query.Get("match")

		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {

			ok, err := match(isMatch, string(value))
			if err != nil {
				return err
			}

			if ok {
				if skip != 0 {
					skip--
					continue
				}
				res = string(value)
				return nil
			}
		}

		return nil
	})

	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}

	if res != "" {
		return res
	}

	return `{"status":"nothing match"}`
}

// TODO updateOne updates one  document data
func (db *Store) updateOne(query gjson.Result) (result string) {

	isMatch := query.Get("match")

	newObj := query.Get("data").Str
	if newObj == "" {
		return `{"error":"no data to update"}`
	}

	coll := query.Get("collection").Str

	// bbolt
	err := db.db.View(func(tx *bbolt.Tx) error {

		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", "myBucket")
		}
		// Use a cursor to iterate over all key-value pairs in the bucket.
		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {

			ok, err := match(isMatch, string(value))
			if err != nil {
				return err
			}

			if ok {
				newData := gjson.Get(`[`+string(value)+`,`+newObj+`]`, `@join`).Raw
				bucket.Put(key, []byte(newData))
				return nil
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println(err.Error())
		// TODO
		return `{"error": "` + err.Error() + `"}`
	}

	return `{"update:": "done"}`
}

// TODO updateMany update document data
func (db *Store) updateMany(query gjson.Result) (result string) {

	isMatch := query.Get("match")

	newObj := query.Get("data").Raw

	coll := query.Get("collection").Str

	// updates exist value

	// bbolt
	err := db.db.View(func(tx *bbolt.Tx) error {

		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", "myBucket")
		}
		// Use a cursor to iterate over all key-value pairs in the bucket.
		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {

			ok, err := match(isMatch, string(value))
			if err != nil {
				return err
			}

			if ok {
				newData := gjson.Get(`[`+string(value)+`,`+newObj+`]`, `@join`).Raw
				bucket.Put(key, []byte(newData))
			}
		}

		return nil
	})

	if err != nil {

		return `{"error": "` + err.Error() + `"}`
	}

	return "many items updated"
}

// deletes Many items
func (db *Store) deleteMany(query gjson.Result) string {

	isMatch := query.Get("match")

	coll := query.Get("collection").Str

	err := db.db.View(func(tx *bbolt.Tx) error {

		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", "myBucket")
		}
		// Use a cursor to iterate over all key-value pairs in the bucket.
		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {

			ok, err := match(isMatch, string(value))
			if err != nil {
				return err
			}

			if ok {
				bucket.Delete(key)
			}
		}

		return nil
	})
	if err != nil {
		return `{"error": "` + err.Error() + `"}`
	}

	return " many items has removed"
}

// Update update document data
func (db *Store) updateById(query gjson.Result) (result string) {

	oldObj := db.findById(query)
	fmt.Println("query ", query)

	id := query.Get("_id").String()
	if id == "" {
		return `{"error": "forget _id"}`
	}
	newObj := query.Get("data").Raw

	coll := query.Get("collection").Str

	newData := gjson.Get(`[`+oldObj+`,`+newObj+`]`, `@join`).Raw
	err := db.db.View(func(tx *bbolt.Tx) error {

		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("collection %q not found", coll)
		}

		bucket.Put([]byte(id), []byte(newData))

		return nil
	})

	if err != nil {
		return `{"error": "` + err.Error() + `"}`
	}

	return id + " updated"
}

// delete one item
func (db *Store) deleteOne(query gjson.Result) string {

	coll := query.Get("collection").Str

	isMatch := query.Get("match")

	id := ""
	err := db.db.View(func(tx *bbolt.Tx) error {

		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", "myBucket")
		}
		// Use a cursor to iterate over all key-value pairs in the bucket.
		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {

			ok, err := match(isMatch, string(value))
			if err != nil {
				return err
			}

			if ok {
				id = string(key)
				bucket.Delete(key)
			}
		}

		return nil
	})
	if err != nil {
		return `{"error": "` + err.Error() + `"}`
	}
	return `{"result":"_id:` + id + ` deleted"}`

}

// Find finds any obs match creteria.
func (db *Store) findMany(query gjson.Result) (res string) {

	listData, err := db.getData(query)
	if err != nil {
		return err.Error()
	}

	// order :
	srt := query.Get("sort")

	if srt.Exists() {
		listData = order(listData, srt)
	}

	// remove or rename some fields
	flds := query.Get("fields")
	listData = reFields(listData, flds)

	records := "["

	for i := 0; i < len(listData); i++ {
		records += listData[i] + ","
	}

	ln := len(records)
	if ln == 1 {
		return "[]"
	}

	return records[:ln-1] + "]"
}

// Finds first obj match creteria.
func (db *Store) findById(query gjson.Result) (res string) {
	coll := query.Get("collection").Str
	key := query.Get("_id").String()

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("collection %s not exist", coll)
		}

		res = string(bucket.Get([]byte(key)))

		return nil
	})
	if err != nil {
		return `{"error": "` + err.Error() + `"}`
	}

	return res
}

// Insert One
func (db *Store) insertOne(query gjson.Result) (res string) {

	coll := query.Get("collection").Str
	if coll == "" {
		return `{"error":"forgot collection"}`
	}
	data := query.Get("data").String() // .Str not works with json obj

	if data == "" {
		return `{"error":"forgot data"}`
	}

	db.lastid[coll]++
	key := strconv.Itoa(int(db.lastid[coll]))
	data = `{"_id":` + key + ", " + data[1:]

	err := db.Put(coll, key, data)
	if err != nil {
		fmt.Println(err)
		db.lastid[collection]--
		return err.Error()
	}

	return `{"ak":"insert ` + key + ` success"}`
}

// InsertMany inserts list of object at one time
func (db *Store) insertMany(query gjson.Result) (res string) {
	coll := query.Get("collection").Str
	data := query.Get("data").Array()

	for _, obj := range data {
		db.lastid[coll]++
		// strconv for perf
		key := strconv.Itoa(int(db.lastid[coll]))
		obj := `{"_id":` + key + `,` + obj.String()[1:]

		err := db.Put(coll, key, obj)
		if err != nil {
			fmt.Println(err)
			db.lastid[collection]--
			return err.Error()
		}

	}

	return `{"ak":"insertMany Done"}`
}

// delete by id
func (db *Store) deleteById(query gjson.Result) string {
	id := query.Get("_id").String()
	if id == "" {
		return `{"error": "_id is required"}`
	}
	coll := query.Get("collection").Str

	err := db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(coll))
		if bucket == nil {
			return fmt.Errorf("%s", `{"error": "collection is required"}`)
		}

		bucket.Delete([]byte(id))

		return nil
	})
	if err != nil {
		return `{"error": "internal error"}` // + err.Error()
	}

	return `{"aknowlge": "row ` + id + ` deleted"}`
}

// end
