package logparser

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

func PrintError(err error, depth int) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(depth)
		filename := path.Base(fn)
		log.Printf("[error] in %s:%d : %v", filename, line, err)
	}
}

func Check(err error) {
	if err != nil {
		PrintError(err, 2)
		os.Exit(1)
	}
}

func FileToLines(path string) ([]string, error) {
	file, err := os.Open(path)
	Check(err)
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// GetStringValue returns a string from the input line located between the beacons.
func GetStringValue(leftBeacon, rightBeacon, line string) (string, error) {
	left := strings.Index(line, leftBeacon)
	if left == -1 {
		return "", errors.New("left beacon not found")
	}
	index := left + len(leftBeacon)
	right := strings.Index(line[index:], rightBeacon)
	if right == -1 {
		return "", errors.New("right beacon not found")
	}
	return line[index : index+right], nil
}

func GetMapValues(path string) (*map[string]float64, error) {
	lines, err := FileToLines(path)
	Check(err)
	m := make(map[string]float64)
	for _, v := range lines {
		val, err := GetStringValue("[[", "]]", v)
		Check(err)
		intVal, err := strconv.ParseFloat(val, 64)
		Check(err)
		m[val] += intVal
	}
	return &m, nil
}
