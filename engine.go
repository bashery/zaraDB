package main

import (
	"fmt"
	"io"
	"os"

	"github.com/tidwall/gjson"
)

// TODO check if named returns improves performence ?!

// Opendb opens | create new file
func Opendbs(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	return file, err
}

// GetVal returns data as string.
// it take file pointr, at int64 & len of data that will read
func GetVal(file *os.File, at int64, buff int) string {
	// TODO check if reusing global buffer fast !
	buffer := make([]byte, buff)

	// read at
	n, err := file.ReadAt(buffer, at)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return ""
	}
	// out the buffer content
	return string(buffer[:n])
}

// AppendData appends data to file
// return lenght of file or err
func AppendData(file *os.File, data string) (int, error) {
	lenByte, err := file.WriteString(data)
	return lenByte, err
}

// getField get field from json string
func getField(field, json string) string {
	return gjson.Get(json, field).String()
}
