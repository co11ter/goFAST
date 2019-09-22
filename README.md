goFAST
======

[![GoDoc](https://godoc.org/github.com/co11ter/goFAST?status.svg)](https://godoc.org/github.com/co11ter/goFAST)

goFAST is a Go implementation of the FAST Protocol (FIX Adapted for STreaming).
Work-in-progress, expect bugs and missing features.

Installation
------------

Install the FAST library using the "go get" command:

    go get github.com/co11ter/goFAST

Usage
-----

See [documentation](https://godoc.org/github.com/co11ter/goFAST) examples.

Benchmark
---------
Run `go test -bench=.`. Only Decoder Benchmark is implemented.

    $ go test -bench=.
    goos: linux
    goarch: amd64
    pkg: github.com/co11ter/goFAST
    BenchmarkDecoder_DecodeReflection-4   	  200000	     10403 ns/op	     795 B/op	      68 allocs/op
    BenchmarkDecoder_DecodeReceiver-4     	  300000	      5453 ns/op	     321 B/op	      32 allocs/op
    PASS
    ok  	github.com/co11ter/goFAST	4.977s
    
TODO
----

- apply errors
- add benchmark of encoder
- optimize encoder
- implement delta and tail operators for string type