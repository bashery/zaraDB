package dblite

import (
	"fmt"
	"io"
	"os"

	"github.com/Jeffail/gabs/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// data enginge

// At is where enginge insert data in page
var At int

var MaxObjects int64 = 10_000

const slash = string(os.PathSeparator) // not tested for windos

// Update update document data
func Update(query string) (result string) {
	collection := gjson.Get(query, "in").String() + slash
	if len(collection) == 1 {
		return "ERROR! select no collection "
	}

	data := SelectById(query)
	newData := gjson.Get(query, "data").String()

	// `{"object":{"first":1,"second":2,"third":3}}`
	jsonParsed, err := gabs.ParseJSON([]byte(newData))
	if err != nil {
		return fmt.Sprintf("ERROR: parse data json %s", err)
	}

	// extract fields that need to update
	for field, val := range jsonParsed.ChildrenMap() {
		result, _ = sjson.Set(data, field, val)
		data = result
	}

	id := gjson.Get(data, "_id").Int()

	path := db.Name + collection + fmt.Sprint(id/MaxObjects)

	_, err = Append(db.Pages[path], data)
	if err != nil {
		return fmt.Errorf("ERROR! from Append %v\n", err).Error()
	}

	// Update index
	size := len(data)

	UpdateIndex(db.Pages[db.Name+collection+pi], int(id), int64(At), int64(size))

	At += size

	return "Success update"
}

// Insert
func Insert(query string) (res string) {

	collection := gjson.Get(query, "in").String() + slash
	if len(collection) == 1 {
		return fmt.Sprint("failure insert. insert into no collection")
	}

	data := gjson.Get(query, "data").String()

	value, err := sjson.Set(data, "_id", db.PrimaryIndex)
	if err != nil {
		fmt.Println("sjson.Set : ", err)
	}

	full := db.PrimaryIndex / MaxObjects

	if full != 0 {
		pageName := db.Name + collection + fmt.Sprint(full)
		//iLog.Println("path in new page is ", pageName)

		page, err := os.OpenFile(pageName, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Println("os open file: ", err)
		}
		db.Pages[pageName] = page
	}

	path := db.Name + collection + fmt.Sprint(db.PrimaryIndex/MaxObjects)

	size, err := Append(db.Pages[path], value)
	if err != nil {
		eLog.Printf("%v Path is %s ", err, path)
		iLog.Println("file page is ", db.Pages[path])
		return "Fielure Insert"
	}

	// set new index
	NewIndex(db.Pages[db.Name+collection+pi], At, len(value))
	At += size
	db.PrimaryIndex++
	return fmt.Sprint("Success Insert, _id: ", db.PrimaryIndex-1)
}

func selectFields(query string) string {

	return ""
}

// delete
func DeleteById(query string) (result string) {

	id := gjson.Get(query, "_id").Int()
	in := gjson.Get(query, "in").String() + slash

	path := db.Name + in + fmt.Sprint(db.PrimaryIndex/MaxObjects)

	fmt.Println("path id DeleteById: ", path)

	UpdateIndex(db.Pages[path], int(id), 0, 0)

	//fmt.Println(IndexsCache.indexs)
	IndexsCache.indexs[id] = [2]int64{0, 0}
	//fmt.Println(IndexsCache.indexs)

	return "Delete Success!"
}

// Select reads data form docs
func SelectById(query string) (result string) {
	id := gjson.Get(query, "where_id").Int()
	if int(id) >= len(IndexsCache.indexs) {
		iLog.Println("no found index")
		return fmt.Sprintf("Not Found _id %v\n", id)
	}

	at := IndexsCache.indexs[id][0]
	size := IndexsCache.indexs[id][1]

	in := gjson.Get(query, "in").String() + slash
	//fmt.Println("table is : ", in)
	// TODO check from if exist!

	path := db.Name + in + fmt.Sprintf("%d", id/MaxObjects)

	result = Get(db.Pages[path], at, int(size))

	return result
}

// appends data to Pagefile & returns file size or error
func Append(file *os.File, data string) (size int, err error) {
	size, err = file.WriteAt([]byte(data), int64(At))
	if err != nil {
		eLog.Println("Error WriteString ", err)
	}
	return size, err
}

// Select reads data form docs
func Select(filter string) (result string) {
	id := gjson.Get(filter, "_id").String()
	fmt.Println("id is ", id)

	return result
}

// gets data from *file, takes at (location) & buffer size
func Get(file *os.File, at int64, size int) string {

	buffer := make([]byte, size)

	// read at
	n, err := file.ReadAt(buffer, at)
	if err != nil && err != io.EOF {
		eLog.Println(err)
		eLog.Println("At", at)
		eLog.Println("Size", size)
		return "ERROR form Get::ReadAt func"
	}

	// out the buffer content
	return string(buffer[:n])
}

// Delete removes document
func Delete(path string) (err error) {
	return
}

// wht is fast ? remander or divider ? 3000/10 or 3000 % 10. for speed during extract dataPage form id
