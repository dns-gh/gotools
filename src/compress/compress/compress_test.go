package compress

import (
	"gotrace"
	"os"
	"path/filepath"
	"testing"
)

const (
	currentFolder = "."
	fileTest1     = "test.1"
	fileTest2     = "test.2"
	testData      = "this is a test"
)

func makeFiles(t *testing.T) string {
	unique := gotrace.MakeUniqueFolder(t)
	root := filepath.Join(currentFolder, unique)
	sub := filepath.Join(currentFolder, unique, gotrace.MakeUniqueFolder(t))
	os.MkdirAll(sub, 0777)
	file1, err := os.Create(filepath.Join(root, fileTest1))
	gotrace.Assert(t, err)
	defer file1.Close()
	file2, err := os.Create(filepath.Join(sub, fileTest2))
	gotrace.Assert(t, err)
	_, err = file2.WriteString(testData)
	gotrace.Assert(t, err)
	defer file2.Close()
	return root
}

func makeZip(t *testing.T) string {
	root := makeFiles(t)
	zipFile, err := Zip(root)
	gotrace.Assert(t, err)
	info, err := os.Stat(zipFile)
	gotrace.Assert(t, err)
	gotrace.Check(t, info.Name() == zipFile)
	gotrace.Check(t, info.Size() > 0)
	err = os.RemoveAll(root)
	gotrace.Assert(t, err)
	return zipFile
}

func TestZip(t *testing.T) {
	zipFile := makeZip(t)
	err := os.RemoveAll(zipFile)
	gotrace.Assert(t, err)
}

func TestUnzip(t *testing.T) {
	zipFile := makeZip(t)
	dst, err := Unzip(zipFile)
	gotrace.Assert(t, err)
	dst1, err := Unzip(zipFile)
	gotrace.Assert(t, err)
	dst2, err := Unzip(zipFile)
	gotrace.Assert(t, err)
	gotrace.Check(t, dst1 == dst+".0")
	gotrace.Check(t, dst2 == dst+".1")
	err = os.RemoveAll(dst1)
	gotrace.Assert(t, err)
	err = os.RemoveAll(dst2)
	gotrace.Assert(t, err)
	num := 0
	found := 0
	filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		num++
		base := filepath.Base(path)
		if base == fileTest1 {
			found++
			gotrace.CheckContent(t, path, "")
		}
		if base == fileTest2 {
			found++
			gotrace.CheckContent(t, path, testData)
		}
		return nil
	})
	gotrace.Check(t, num == 4)
	gotrace.Check(t, found == 2)
	err = os.RemoveAll(zipFile)
	gotrace.Assert(t, err)
	err = os.RemoveAll(dst)
	gotrace.Assert(t, err)
}
