package gotrace

import (
	"crypto/rand"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

const (
	testDir = "test"
)

func traceBack(t *testing.T, err error) {
	info := ""
	if err != nil {
		info = err.Error()
	}
	t.Error("Test failed ", info)
	runtime.Goexit()
}

// Assert makes the test to fail and dumps
// the trace if error is not nil
func Assert(t *testing.T, err error) {
	if err != nil {
		traceBack(t, err)
	}
}

// Check checks that the condition is verified.
// It makes the test to fail and dumps the trace
// as well otherwise.
func Check(t *testing.T, condition bool) {
	if !condition {
		traceBack(t, nil)
	}
}

// CheckContent checks that the input file contains
// the expected information.
// It makes the test to fail and dumps the trace
// as well otherwise.
func CheckContent(t *testing.T, path, expected string) {
	file, err := os.Open(path)
	Assert(t, err)
	defer file.Close()
	bytes := make([]byte, len(expected))
	_, err = file.Read(bytes)
	Assert(t, err)
	Check(t, string(bytes) == expected)
}

// MakeUniqueFolder creates a folder with pseudo
// random generator
func MakeUniqueFolder(t *testing.T) string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	Assert(t, err)
	return strconv.FormatUint(binary.BigEndian.Uint64(bytes), 10)
}

// MakeUniqueTestFolder creates a unique test folder inside
// the pre-defined test folder.
func MakeUniqueTestFolder(t *testing.T) string {
	return filepath.Join(testDir, MakeUniqueFolder(t))
}

// GetTestFolder returns the hardcoded test folder
func GetTestFolder() string {
	return testDir
}

// RemoveTestFolder removes the hardcoded test folder
func RemoveTestFolder(t *testing.T) {
	if !t.Failed() {
		Assert(t, os.RemoveAll(GetTestFolder()))
	}
}
