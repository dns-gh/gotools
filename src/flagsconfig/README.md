## flagsconfig package

[![GoDoc](https://godoc.org/github.com/dns-gh/gotools/src/flagsconfig?status.png)]
(https://godoc.org/github.com/dns-gh/gotools/src/flagsconfig)

Improved configuration management using a configuration file with key-value pair in json format overloaded by user defined flags and used flags at runtime.

## Motivation

To get a more clean and simple way to play with configuration files.

## Installation

- It requires Go language of course. You can set it up by downloading it here: https://golang.org/dl/
- Use go get or download the files directly from github to get the project
- Set your GOPATH (to the project location) and GOROOT (where Go is installed) environment variables.

## Build

```
@gotools $ go install flagsconfig
```

## Example

File: testConfig.go
```
package main

import (
	"flag"
	"flagsconfig"
	"fmt"
)

func main() {
	filename := "test.config"
	firstFlag := flag.String("first", "firstDefault", "first user defined flag")
	otherFlag := flag.String("other", "otherDefault", "another user defined flag")

	// Makes the configuration structure with one filtered flag being the config flag
	// Hence, it will not be saved in the configuration file.
	// And parses the user defined flags and the one used by the user at runtime.
	// If the 'first' flag was used at runtime, the value of this flag will be
	// the one saved into the config file instead of the default one.
	_, err := flagsconfig.NewConfig(filename, "other")
	if err != nil {
		panic(err)
	}

	fmt.Println("filename:", filename)
	fmt.Println("first:", *firstFlag)
	fmt.Println("other:", *otherFlag)
}
```

So with
```
$ bin/testConfig.exe
```

we get the following output

```
filename: test.config
first: firstDefault
other: otherDefault
```

and the configuration file looks like:

```
{
    "first": "firstDefault"
}
```

but with
```
$ bin/testConfig.exe -first "alice"
```

we get the following output

```
filename: test.config
first: alice
other: otherDefault
```

and the configuration file looks like:

```
{
    "first": "alice"
}
```

if we run one more time
```
$ bin/testConfig.exe
```

we get
```
filename: test.config
first: alice
other: otherDefault
```

and the configuration file looks like:

```
{
    "first": "alice"
}
```
since the 'first' flag was not defined, the previous value contained in the config file was used.

## Tests

```
@gotools $ go test -v flagsconfig
=== RUN   TestFlagsConfig
--- PASS: TestFlagsConfig (0.00s)
=== RUN   TestFlagsConfigFiltered
--- PASS: TestFlagsConfigFiltered (0.00s)
PASS
ok      flagsconfig     0.058s
```