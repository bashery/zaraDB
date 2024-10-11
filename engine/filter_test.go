package engine

import (
	"testing"

	"github.com/tidwall/gjson"
)

var testCases = []struct {
	filter   string
	data     string
	expected bool
}{
	// ....filter.....     ...data...    ...expected result...
	// boalans & null cases
	{`{"graded":true}`, `{"graded":true}`, true},
	{`{"graded":true}`, `{"graded":false}`, false},

	{`{"graded":false}`, `{"graded":true}`, false},
	{`{"graded":false}`, `{"graded":false}`, true},

	{`{"graded":null}`, `{"graded":null}`, true},
	{`{"graded":null}`, `{"graded":true}`, false},
	{`{"graded":null}`, `{"graded":"null"}`, false},

	//  cases of numbers
	{`{"age": 18}`, `{"age": 18}`, true},
	{`{"age": 18}`, `{"age": 19}`, false},

	{`{"age":{"$eq":18}}`, `{"age": 18}`, true},
	{`{"age":{"$eq":18}}`, `{"age": 19}`, false},

	{`{"age":{"$ne": 18}}`, `{"age": 19}`, true},
	{`{"age":{"$ne": 18}}`, `{"age": 18}`, false},

	{`{"age":{"$gt": 18}}`, `{"age": 19}`, true},
	{`{"age":{"$gt": 18}}`, `{"age": 18}`, false},
	{`{"age":{"$gt": 18}}`, `{"age": 17}`, false},

	{`{"age":{"$lt": 18}}`, `{"age": 17}`, true},
	{`{"age":{"$lt": 18}}`, `{"age": 18}`, false},
	{`{"age":{"$lt": 18}}`, `{"age": 19}`, false},

	{`{"age":{"$gte": 18}}`, `{"age": 19}`, true},
	{`{"age":{"$gte": 18}}`, `{"age": 18}`, true},
	{`{"age":{"$gte": 18}}`, `{"age": 17}`, false},

	{`{"age":{"$lte": 18}}`, `{"age": 17}`, true},
	{`{"age":{"$lte": 18}}`, `{"age": 18}`, true},
	{`{"age":{"$lte": 18}}`, `{"age": 19}`, false},

	{`{"age":{"$lt": 28, "$gt": 18}}`, `{"age": 20}`, true},
	{`{"age":{"$lt": 28, "$gt": 18}}`, `{"age": 27}`, true},
	{`{"age":{"$lt": 28, "$gt": 18}}`, `{"age": 19}`, true},
	{`{"age":{"$lt": 28, "$gt": 18}}`, `{"age": 18}`, false},

	{`{"age":{"$lte": 28, "$gte": 18}}`, `{"age": 20}`, true},
	{`{"age":{"$lte": 28, "$gte": 18}}`, `{"age": 28}`, true},
	{`{"age":{"$lte": 28, "$gte": 18}}`, `{"age": 18}`, true},
	{`{"age":{"$lte": 28, "$gte": 18}}`, `{"age": 16}`, false},

	{`{"age":{"$in":[28,29,30]}}`, `{"age": 29}`, true},
	{`{"age":{"$in":[28,29,30]}}`, `{"age": 9}`, false},

	{`{"age":{"$nin":[28,29,30]}}`, `{"age": 9}`, true},
	{`{"age":{"$nin":[28,29,30]}}`, `{"age": 29}`, false},

	{`{"$and":[{"name":{"$eq":"adam"}}, {"age":{"$eq":29}} ] }`, `{"name":"adam","age": 29}`, true},
	{`{"$and":[{"name":{"$eq":"adam"}}, {"age":{"$eq":29}} ] }`, `{"name":"adam","age": 19}`, false},

	{`{"$or":[{"name":{"$eq":"adam"}}, {"age":{"$eq":29}} ] }`, `{"name":"jhon","age": 29}`, true},
	{`{"$or":[{"name":{"$eq":"adam"}}, {"age":{"$eq":29}} ] }`, `{"name":"jhon","age": 19}`, false},

	// string cases
	{`{"name":"adam"}`, `{"name":"adam"}`, true},
	{`{"name":"adam"}`, `{"name":"kamal"}`, false},

	{`{"name":{"$eq":"adam"}}`, `{"name":"adam"}`, true},
	{`{"name":{"$eq":"adam"}}`, `{"name":"john"}`, false},

	{`{"name":{"$ne":"adam"}}`, `{"name":"john"}`, true},
	{`{"name":{"$ne":"adam"}}`, `{"name":"adam"}`, false},

	{`{"name":{"$st":"ad"}}`, `{"name":"adam"}`, true},
	{`{"name":{"$st":"ad"}}`, `{"name":"john"}`, false},

	{`{"name":{"$nst":"ad"}}`, `{"name":"john"}`, true},
	{`{"name":{"$nst":"ad"}}`, `{"name":"adam"}`, false},

	{`{"name":{"$en":"am"}}`, `{"name":"adam"}`, true},
	{`{"name":{"$en":"ad"}}`, `{"name":"john"}`, false},

	{`{"name":{"$nen":"hn"}}`, `{"name":"adam"}`, true},
	{`{"name":{"$nen":"am"}}`, `{"name":"adam"}`, false},
}

func Test_Match(t *testing.T) {

	for _, tcase := range testCases {
		filt := gjson.Parse(tcase.filter)
		result, _ := match(filt, tcase.data)

		if result != tcase.expected {
			t.Errorf(Yellow+"\nfilter:  %s\ndata:    %s\nexpected %v,\ngot:     %v\n"+Reset,
				tcase.filter, tcase.data, tcase.expected, result)
		}
	}
}

func assert(t *testing.T, filt, data string, exp bool) {
	parsed := gjson.Parse(filt)
	result, _ := match(parsed, data)

	if result != exp {
		t.Errorf(
			Yellow+"\nfilter:  %s\ndata:    %s\nexpected %v,\ngot:     %v\n"+
				Reset, filt, data, exp, result)
	}
}

var data = []string{

	`{"_id": 2, "name":"karim", "age":37, "contact":{"email": "karim@email.com", "tele": "00234856781"}}`,
	`{"_id": 3, "name":"jawad", "age":24, "contact":{"email": "jawad@email.com", "tele": "00123845672"}}`,
	`{"_id": 4, "name":"imane", "age":26, "contact":{"email": "imane@email.com", "tele": "00239850303"}}`,
	`{"_id": 5, "name":"adams", "age":31, "contact":{"email": "adams@email.com", "tele": "00234850004"}}`,
	`{"_id": 6, "name":"johns", "age":11, "contact":{"email": "adams@email.com", "tele": "00234850105"}}`,
	`{"_id": 7, "name":"jawar", "age":10, "contact":{"email": "jawar@email.com", "tele": "00234850226"}}`,
	`{"_id": 8, "name":"kadir", "age":13, "contact":{"email": "kadir@email.com", "tele": "00238450337"}}`,
	`{"_id": 9, "name":"hamid", "age":16, "contact":{"email": "hamid@email.com", "tele": "00238450048"}}`,
	`{"_id":10, "name":"rajab", "age":18, "contact":{"email": "rajab@email.com", "tele": "00230450059"}}`,
	`{"_id":11, "name":"samir", "age":19, "contact":{"email": "samir@email.com", "tele": "00230450060"}}`,
	`{"_id":12, "name":"akram", "age":21, "contact":{"email": "akram@email.com", "tele": "00239450071"}}`,
	`{"_id":13, "name":"rabih", "age":31, "contact":{"email": "rabih@email.com", "tele": "00230450083"}}`,
	`{"_id":14, "name":"monir", "age":31, "contact":{"email": "monir@email.com", "tele": "00239450094"}}`,
}

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
)
