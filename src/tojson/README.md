## tojson package

Simple wrapper over the encoding/json package to save and load data to a file into json format

## Motivation

Used in the flagsconfig package.

## Installation

- It requires Go language of course. You can set it up by downloading it here: https://golang.org/dl/
- Use go get or download the files directly from github to get the project
- Set your GOPATH (to the project location) and GOROOT (where Go is installed) environment variables.

## Build

```
@gotools $ go install tojson
```

## Usage

```
package main

import "tojson"

type equipment struct {
	Weapon string `json:"weapon"`
	Shield string `json:"shield"`
}

func main() {
	oldEquipment := &equipment{
		Weapon: "sword",
		Shield: "wooden",
	}
	file := "equipment.json"
	tojson.Save(file, oldEquipment)
	newEquipement := &equipment{}
	tojson.Load(file, newEquipement)
}

```

## Tests

```
@gotools $ go test -v tojson
=== RUN   TestSaveLoadJSON
--- PASS: TestSaveLoadJSON (0.00s)
PASS
ok      tojson  0.055s
```