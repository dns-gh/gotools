package flagsconfig

import (
	"flag"
	"gotrace"
	"os"
	"path/filepath"
	"testing"
)

const (
	configFileName       = "test.config"
	flagTest             = "flag-test"
	flagTestValue        = "test"
	flagTestDescription  = "configuration test flag"
	flagTest2            = "flag-test2"
	flagTestValue2       = "test2"
	flagTestDescription2 = "configuration test flag 2"
	flagTest3            = "flag-test3"
	flagTestValue3       = "test3"
	flagTestDescription3 = "configuration test flag 3"
)

func makeConfigFile(dir string) string {
	return filepath.Join(dir, configFileName)
}

func checkFileInfo(t *testing.T, file string) {
	info, err := os.Stat(file)
	gotrace.Assert(t, err)
	gotrace.Check(t, info.Name() == configFileName)
	gotrace.Check(t, info.Size() != 0)
	gotrace.Check(t, !info.IsDir())
}

func checkFlag(t *testing.T, flagTest, flagTestValue string) {
	testFlag := flag.Lookup(flagTest)
	gotrace.Check(t, testFlag.Value.String() == flagTestValue)
}

func TestFlagsConfig(t *testing.T) {
	defer gotrace.RemoveTestFolder(t)
	flag.String(flagTest, "", flagTestDescription)

	dir := gotrace.MakeUniqueTestFolder(t)
	file := makeConfigFile(dir)
	config, err := NewConfig(file)
	gotrace.Assert(t, err)
	checkFileInfo(t, file)
	checkFlag(t, flagTest, "")

	err = config.Update(flagTest, flagTestValue)
	gotrace.Assert(t, err)
	checkFlag(t, flagTest, "")

	err = config.Parse(file)
	gotrace.Assert(t, err)
	checkFlag(t, flagTest, flagTestValue)
}

func TestFlagsConfigFiltered(t *testing.T) {
	defer gotrace.RemoveTestFolder(t)
	flag.String(flagTest2, "", flagTestDescription2)
	flag.String(flagTest3, "", flagTestDescription3)

	dir := gotrace.MakeUniqueTestFolder(t)
	file := makeConfigFile(dir)
	config, err := NewConfig(file, flagTest3)
	gotrace.Assert(t, err)
	checkFileInfo(t, file)

	checkFlag(t, flagTest2, "")
	checkFlag(t, flagTest3, "")

	err = config.Update(flagTest2, flagTestValue2)
	gotrace.Assert(t, err)

	checkFlag(t, flagTest2, "")
	checkFlag(t, flagTest3, "")

	err = config.Parse(file)
	gotrace.Assert(t, err)

	checkFlag(t, flagTest2, flagTestValue2)
	checkFlag(t, flagTest3, "")
}
