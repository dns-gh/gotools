package tojson

import (
	"gotrace"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	fileTest = "test.test"
)

type jsonTest struct {
	Field1 string            `json:"field1"`
	Field2 int               `json:"field2"`
	Field3 bool              `json:"field3"`
	Field4 map[string]string `json:"field4"`
}

func TestSaveLoadJSON(t *testing.T) {
	defer gotrace.RemoveTestFolder(t)
	expected := &jsonTest{
		Field1: "test",
		Field2: 2,
		Field3: true,
		Field4: map[string]string{
			"test1": "val1",
			"test2": "val2",
		},
	}
	current := &jsonTest{}
	file := filepath.Join(gotrace.GetTestFolder(), fileTest)
	gotrace.Assert(t, Save(file, expected))
	gotrace.Assert(t, Load(file, current))
	gotrace.Check(t, reflect.DeepEqual(expected, current))
}
