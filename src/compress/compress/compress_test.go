package compress

import (
	"crypto/rand"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

const (
	removeTimeout = 15 * time.Second
	currentFolder = "."
	fileTest1     = "test.1"
	fileTest2     = "test.2"
	testData      = "this is a test"
)

func traceBack(t *testing.T, err error) {
	info := ""
	if err != nil {
		info = err.Error()
	}
	t.Error("Test failed ", info)
	runtime.Goexit()
}

func assert(t *testing.T, err error) {
	if err != nil {
		traceBack(t, err)
	}
}

func check(t *testing.T, condition bool) {
	if !condition {
		traceBack(t, nil)
	}
}

// makeUniqueFolder creates a folder with pseudo
// random generator
func makeUniqueFolder(t *testing.T) string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		t.Error("Test failed:", err.Error())
	}
	return strconv.FormatUint(binary.BigEndian.Uint64(bytes), 10)
}

func makeFiles(t *testing.T) string {
	unique := makeUniqueFolder(t)
	root := filepath.Join(currentFolder, unique)
	sub := filepath.Join(currentFolder, unique, makeUniqueFolder(t))
	os.MkdirAll(sub, 0777)
	file1, err := os.Create(filepath.Join(root, fileTest1))
	assert(t, err)
	defer file1.Close()
	file2, err := os.Create(filepath.Join(sub, fileTest2))
	assert(t, err)
	_, err = file2.WriteString(testData)
	assert(t, err)
	defer file2.Close()
	return root
}

func makeZip(t *testing.T) string {
	root := makeFiles(t)
	zipFile, err := Zip(root)
	assert(t, err)
	info, err := os.Stat(zipFile)
	assert(t, err)
	check(t, info.Name() == zipFile)
	check(t, info.Size() > 0)
	err = os.RemoveAll(root)
	assert(t, err)
	return zipFile
}

func TestZip(t *testing.T) {
	zipFile := makeZip(t)
	err := os.RemoveAll(zipFile)
	assert(t, err)
}

func checkContent(t *testing.T, path, expected string) {
	file, err := os.Open(path)
	assert(t, err)
	defer file.Close()
	bytes := make([]byte, len(expected))
	_, err = file.Read(bytes)
	assert(t, err)
	check(t, string(bytes) == expected)
}

func TestUnzip(t *testing.T) {
	zipFile := makeZip(t)
	dst, err := Unzip(zipFile)
	assert(t, err)
	dst1, err := Unzip(zipFile)
	assert(t, err)
	dst2, err := Unzip(zipFile)
	assert(t, err)
	check(t, dst1 == dst+".0")
	check(t, dst2 == dst+".1")
	err = os.RemoveAll(dst1)
	assert(t, err)
	err = os.RemoveAll(dst2)
	assert(t, err)
	num := 0
	found := 0
	filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		num += 1
		base := filepath.Base(path)
		if base == fileTest1 {
			found += 1
			checkContent(t, path, "")
		}
		if base == fileTest2 {
			found += 1
			checkContent(t, path, testData)
		}
		return nil
	})
	check(t, num == 4)
	check(t, found == 2)
	err = os.RemoveAll(zipFile)
	assert(t, err)
	err = os.RemoveAll(dst)
	assert(t, err)
}
