package logparser

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
)

func CheckBool(value bool) {
	if !value {
		PrintError(errors.New("not the whole truth !"), 2)
		os.Exit(1)
	}
}

func CheckStrings(lhs, rhs string) {
	if len(lhs) != len(rhs) {
		PrintError(errors.New(lhs+" != "+rhs), 2)
		os.Exit(1)
	}
	if !strings.Contains(lhs, rhs) {
		PrintError(errors.New(lhs+" != "+rhs), 2)
		os.Exit(1)
	}
}

func CheckInt64(lhs, rhs int64) {
	if lhs != rhs {
		PrintError(errors.New(string(lhs)+" != "+string(rhs)), 2)
		os.Exit(1)
	}
}

func CheckFloat64(lhs, rhs float64) {
	if lhs != rhs {
		PrintError(errors.New(strconv.FormatFloat(lhs, 'f', 6, 64)+" != "+strconv.FormatFloat(rhs, 'f', 6, 64)), 2)
		os.Exit(1)
	}
}

func TestLogParser(t *testing.T) {
	path := "../../../testdata/logdata/test.log"
	lines, err := FileToLines(path)
	Check(err)
	CheckBool(lines[0] == string("message test 1 [[1234]]"))
	CheckStrings(lines[1], string("message test 2 [[12.34]]"))
	value, err := GetStringValue("[[", "]]", "Test Line [[1234]]")
	Check(err)
	CheckStrings(value, string("1234"))

	m, err := GetMapValues(path)
	expectedString := []string{"1234", "12.34"}
	expectedInt := []float64{1234, 12.34}
	var index = 0
	for key, value := range *m {
		CheckStrings(key, expectedString[index])
		CheckFloat64(value, expectedInt[index])
		index++
	}
}
