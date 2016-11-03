// Package tojson allow the user to save and load data of structs
// into json format to a given file.
package tojson

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// Load load json formated data from a given file
func Load(path string, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}

// Save save data in json format into the given file
func Save(path string, data interface{}) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir+0666)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	buf, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	_, err = io.Copy(file, bytes.NewReader(buf))
	return err
}
