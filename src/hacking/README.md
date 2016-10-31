## Hacking tools

First draw of hacking tools. Only basic brute force is available for now.

## Motivation

For fun and by curiosity. Feel free to join my efforts!

## Installation

- It requires Go language of course. You can set it up by downloading it here: https://golang.org/dl/
- Use go get or download the files directly from github to get the project
- Set your GOPATH (to the project location) and GOROOT (where Go is installed) environment variables.

## Build and usage

```
@gotools $ go install hacking/brute
@gotools $ bin/brute.exe -l=3 -c=8 -s="abc"
2016/09/24 23:37:08 length (-l) 3
2016/09/24 23:37:08 cpu (-c) 8
2016/09/24 23:37:08 charset (-s) abc
2016/09/24 23:37:08 running brute force...
2016/09/24 23:37:08 a
2016/09/24 23:37:08 b
2016/09/24 23:37:08 c
2016/09/24 23:37:08 aa
...
2016/09/24 23:37:08 ccc
2016/09/24 23:37:08 brute force ended
```

## Tests and benchmarks

```
@gotools $ go test -v hacking/algorithms
=== RUN   TestBruteForce
--- PASS: TestBruteForce (0.03s)
PASS
ok      hacking/algorithms      0.437s
```

```
@gotools $ go test -v hacking/algorithms -run=XXX -bench=. -benchtime=10s
PASS
BenchmarkBruteForceCPU1-8             20         913067575 ns/op
testing: BenchmarkBruteForceCPU1-8 left GOMAXPROCS set to 1
BenchmarkBruteForceCPU2-8             20         887401615 ns/op
testing: BenchmarkBruteForceCPU2-8 left GOMAXPROCS set to 2
BenchmarkBruteForceCPU3-8             20         889073285 ns/op
testing: BenchmarkBruteForceCPU3-8 left GOMAXPROCS set to 3
BenchmarkBruteForceCPU4-8             20         887030195 ns/op
testing: BenchmarkBruteForceCPU4-8 left GOMAXPROCS set to 4
ok      hacking/algorithms      75.485s
```