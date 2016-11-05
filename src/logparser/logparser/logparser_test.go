package logparser

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
)

func checkBool(value bool) {
	if !value {
		printError(errors.New("not the whole truth !"), 2)
		os.Exit(1)
	}
}

func checkStrings(lhs, rhs string) {
	if len(lhs) != len(rhs) {
		printError(errors.New(lhs+" != "+rhs), 2)
		os.Exit(1)
	}
	if !strings.Contains(lhs, rhs) {
		printError(errors.New(lhs+" != "+rhs), 2)
		os.Exit(1)
	}
}

func checkInt64(lhs, rhs int64) {
	if lhs != rhs {
		printError(errors.New(string(lhs)+" != "+string(rhs)), 2)
		os.Exit(1)
	}
}

func checkFloat64(lhs, rhs float64) {
	if lhs != rhs {
		printError(errors.New(strconv.FormatFloat(lhs, 'f', 6, 64)+" != "+strconv.FormatFloat(rhs, 'f', 6, 64)), 2)
		os.Exit(1)
	}
}

func TestLogParser(t *testing.T) {
	path := "../../../testdata/logdata/test.log"
	lines, err := FileToLines(path)
	check(err)
	checkBool(lines[0] == string("message test 1 [[1234]]"))
	checkStrings(lines[1], string("message test 2 [[12.34]]"))
	value, err := GetStringValue("[[", "]]", "Test Line [[1234]]")
	check(err)
	checkStrings(value, string("1234"))

	m, err := GetMapValues(path)
	check(err)
	expectedString := []string{"1234", "12.34"}
	expectedInt := []float64{1234, 12.34}
	var index = 0
	for key, value := range *m {
		checkStrings(key, expectedString[index])
		checkFloat64(value, expectedInt[index])
		index++
	}
}
