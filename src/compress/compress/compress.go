package compress

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func zipFile(source, path string, writer *zip.Writer) error {
	name, err := filepath.Rel(source, path)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	w, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, file)
	return err
}

func Zip(source string) (string, error) {
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("the specified directory doesn't exist: %s", source)
	} else if err != nil {
		return "", err
	} else if !info.IsDir() {
		return "", fmt.Errorf("the specified input is not a directory: %s", source)
	}
	target := source + ".zip"
	zipped, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer zipped.Close()
	writer := zip.NewWriter(zipped)
	defer writer.Close()
	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		err = zipFile(source, path, writer)
		if err != nil {
			log.Printf("unzipped file '%s' : %v", path, err)
		}
		return nil
	})
	return target, err
}

func makeExt(index int) string {
	return "." + strconv.Itoa(index)
}

// makeFolder returns a custom indexed folder
// if the one provided already exists
func makeFolder(folder string) (string, error) {
	target := strings.TrimSuffix(folder, filepath.Ext(folder))
	index := 0
	supp := makeExt(index)
	previous := target
	for {
		info, err := os.Stat(target)
		if err == nil {
			if info.IsDir() {
				target = previous + supp
				index += 1
				supp = makeExt(index)
			} else {
				break
			}
		} else if os.IsNotExist(err) {
			break
		} else {
			return "", fmt.Errorf("target folder error: %s", err.Error())
		}
	}
	return target, nil
}

func Unzip(archive string) (string, error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	target, err := makeFolder(archive)
	if err != nil {
		return "", err
	}

	if err = os.MkdirAll(target, 0755); err != nil {
		return "", err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		os.MkdirAll(filepath.Dir(path), os.ModeDir)

		fileReader, err := file.Open()
		if err != nil {
			return "", err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return "", err
		}
	}
	return target, nil
}
