# go-fb2parse

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/go-fb2parse.svg)](https://pkg.go.dev/github.com/egnd/go-fb2parse)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/go-fb2parse)](https://goreportcard.com/report/github.com/egnd/go-fb2parse)
[![Coverage](https://gocover.io/_badge/github.com/egnd/go-fb2parse?asdf)](https://gocover.io/github.com/egnd/go-fb2parse)
[![Pipeline](https://github.com/egnd/go-fb2parse/actions/workflows/pipeline.yml/badge.svg)](https://github.com/egnd/go-fb2parse/actions?query=workflow%3APipeline)

Golang package for parsing FB2-files.

## Examples:
```golang
package main

import (
	"encoding/xml"
	"io"

	"github.com/egnd/go-fb2parse"
)

func main() {
	var reader io.Reader
	var fb2Doc fb2parse.FB2File

	// get fb2 struct by xml decoder
	if err := xml.NewDecoder(reader).Decode(&fb2Doc); err != nil {
		panic(err)
	}

	// get fb2 struct by configured xml decoder
	if err := fb2parse.NewDecoder(reader).Decode(&fb2Doc); err != nil {
		panic(err)
	}

	// get fb2 struct by parsing
	if fb2Doc, err := fb2parse.NewFB2File(fb2parse.NewDecoder(reader)); err != nil {
		panic(err)
	}

	// get fb2 struct by parsing withour "binary" images
	if fb2Doc, err := fb2parse.NewFB2File(fb2parse.NewDecoder(reader),
		func(next fb2parse.TokenHandler) fb2parse.TokenHandler {
			return func(obj interface{},
				node xml.StartElement, r xml.TokenReader,
			) (err error) {
				if _, ok := obj.(*fb2parse.FB2File); ok &&
					node.Name.Local == "binary" {
					return SkipToken(node.Name.Local, r)
				}

				return next(obj, node, r)
			}
		},
	); err != nil {
		panic(err)
	}
}
```

## Benchmarks:
```
goos: linux
goarch: amd64
pkg: github.com/egnd/go-fb2parse
cpu: AMD Ryzen 7 5800U with Radeon Graphics         
Benchmark_Decoding/xml-16        42  44711604 ns/op  4723695 B/op   6382 allocs/op
Benchmark_Decoding/fb2-16        43  45216919 ns/op  4723693 B/op   6382 allocs/op
Benchmark_Parsing/rules_0-16     42  29950641 ns/op  2284588 B/op   3975 allocs/op
Benchmark_Parsing/rules_1-16     34  29852034 ns/op  2284509 B/op   3983 allocs/op
Benchmark_Parsing/rules_10-16    43  25313036 ns/op  2285624 B/op   4055 allocs/op
Benchmark_Parsing/rules_100-16   34  31983362 ns/op  2297366 B/op   4774 allocs/op
Benchmark_Parsing/rules_500-16   42  27396962 ns/op  2348383 B/op   7974 allocs/op
Benchmark_Parsing/rules_1000-16  39  30218804 ns/op  2412397 B/op  11975 allocs/op
```