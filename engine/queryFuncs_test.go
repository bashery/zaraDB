package engine

import (
	"testing"

	"github.com/tidwall/gjson"
	"go.etcd.io/bbolt"
)

var s *Store

func Test_NewDB(t *testing.T) {
	s = NewDB("tmptest.db")
	if s == nil {
		t.Error("Store should not be nil")
	}
}

func Test_insertOne(t *testing.T) {
	json := `{"collection":"test", "action":"insertOne","data":{"name":"adam", "age": 23}}`
	query := gjson.Parse(json)
	s.insertOne(query)

	s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("test"))
		res := bucket.Get([]byte("1"))
		exp := []byte(`{"_id":1, "name":"adam", "age": 23}`)
		if string(res) != string(exp) {
			t.Errorf("get %s\nshould be %s", res, exp)
		}

		return nil
	})
}

func Test_findOne(t *testing.T) {
	json := `{"collection":"test", "action":"findOne","match":{"name":"adam"}}` //
	query := gjson.Parse(json)

	res := s.findOne(query)
	exp := []byte(`{"_id":1, "name":"adam", "age": 23}`)

	if string(res) != string(exp) {
		t.Errorf("get %s\nshould be %s", res, exp)
	}
}

func Test_Close(t *testing.T) {
	err := s.db.Close()

	if err != nil {
		t.Error("Store should be nil")
	}
}
