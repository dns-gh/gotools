package compress

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dns-gh/gotest"
)

const (
	fileTest1 = "test.1"
	fileTest2 = "test.2"
	testData  = "this is a test"
)

func makeFiles(t *testing.T) string {
	unique := gotest.MakeUniqueTestFolder(t)
	root := filepath.Join(gotest.GetTestFolder(), unique)
	sub := filepath.Join(gotest.GetTestFolder(), unique, gotest.MakeUniqueFolder(t))
	os.MkdirAll(sub, 0777)
	file1, err := os.Create(filepath.Join(root, fileTest1))
	gotest.Assert(t, err)
	defer file1.Close()
	file2, err := os.Create(filepath.Join(sub, fileTest2))
	gotest.Assert(t, err)
	_, err = file2.WriteString(testData)
	gotest.Assert(t, err)
	defer file2.Close()
	return root
}

func makeZip(t *testing.T) string {
	root := makeFiles(t)
	zipFile, err := Zip(root)
	gotest.Assert(t, err)
	info, err := os.Stat(zipFile)
	gotest.Assert(t, err)
	gotest.Check(t, info.Name() == filepath.Base(zipFile))
	gotest.Check(t, info.Size() > 0)
	err = os.RemoveAll(root)
	gotest.Assert(t, err)
	return zipFile
}

func TestZip(t *testing.T) {
	defer gotest.RemoveTestFolder(t)
	makeZip(t)
}

func TestUnzip(t *testing.T) {
	defer gotest.RemoveTestFolder(t)
	zipFile := makeZip(t)
	dst, err := Unzip(zipFile)
	gotest.Assert(t, err)
	dst1, err := Unzip(zipFile)
	gotest.Assert(t, err)
	dst2, err := Unzip(zipFile)
	gotest.Assert(t, err)
	gotest.Check(t, dst1 == dst+".0")
	gotest.Check(t, dst2 == dst+".1")
	num := 0
	found := 0
	filepath.Walk(dst, func(path string, info os.FileInfo, err error) error {
		num++
		base := filepath.Base(path)
		if base == fileTest1 {
			found++
			gotest.CheckContent(t, path, "")
		}
		if base == fileTest2 {
			found++
			gotest.CheckContent(t, path, testData)
		}
		return nil
	})
	gotest.Check(t, num == 4)
	gotest.Check(t, found == 2)
}
