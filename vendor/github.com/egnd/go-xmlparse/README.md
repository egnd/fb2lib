# go-xmlparse

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/go-xmlparse.svg)](https://pkg.go.dev/github.com/egnd/go-xmlparse)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/go-xmlparse)](https://goreportcard.com/report/github.com/egnd/go-xmlparse)
[![Coverage](https://gocover.io/_badge/github.com/egnd/go-xmlparse?cachefix2)](https://gocover.io/github.com/egnd/go-xmlparse)
[![Pipeline](https://github.com/egnd/go-xmlparse/actions/workflows/pipeline.yml/badge.svg)](https://github.com/egnd/go-xmlparse/actions?query=workflow%3APipeline)

Golang package for parsing xml data.

## Examples:
```golang
package main

import (
	"encoding/xml"
	"io"

	"github.com/egnd/go-xmlparse"
	"github.com/egnd/go-xmlparse/fb2"
)

func main() {
	var reader io.Reader
	var fb2Doc fb2.File

	// get fb2 struct by xml decoder
	if err := xml.NewDecoder(reader).Decode(&fb2Doc); err != nil {
		panic(err)
	}

	// get fb2 struct by configured xml decoder
	if err := xmlparse.NewDecoder(reader).Decode(&fb2Doc); err != nil {
		panic(err)
	}

	// get fb2 struct by parsing
	if fb2Doc, err := fb2.NewFile(xmlparse.NewDecoder(reader)); err != nil {
		panic(err)
	}

	// get fb2 struct by parsing withour "binary" images
	if fb2Doc, err := fb2.NewFile(xmlparse.NewDecoder(reader),
		func(next xmlparse.TokenHandler) xmlparse.TokenHandler {
			return func(obj interface{}, node xml.StartElement, r xmlparse.TokenReader) (err error) {
				if _, ok := obj.(*fb2.File); ok && node.Name.Local == "binary" {
					return xmlparse.TokenSkip(node.Name.Local, r)
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
pkg: github.com/egnd/go-xmlparse
cpu: AMD Ryzen 7 5800U with Radeon Graphics         

Benchmark_Decoders/xml-16                    40   45330854 ns/op   4719552 B/op   6126 allocs/op
Benchmark_Decoders/xmlparse-16               33   43500593 ns/op   4723757 B/op   6390 allocs/op

Benchmark_Parsers_XML/rules_0-16             40   28759416 ns/op   2280436 B/op   3711 allocs/op
Benchmark_Parsers_XML/rules_1-16             55   31546185 ns/op   2280525 B/op   3719 allocs/op
Benchmark_Parsers_XML/rules_10-16            42   27412632 ns/op   2281527 B/op   3791 allocs/op
Benchmark_Parsers_XML/rules_100-16           52   29062773 ns/op   2292956 B/op   4510 allocs/op
Benchmark_Parsers_XML/rules_500-16           39   29684596 ns/op   2344098 B/op   7711 allocs/op
Benchmark_Parsers_XML/rules_1000-16          48   30376356 ns/op   2408053 B/op  11710 allocs/op

Benchmark_Parsers_XMLParsing/rules_0-16      56   30635241 ns/op   2284448 B/op   3975 allocs/op
Benchmark_Parsers_XMLParsing/rules_1-16      55   30133415 ns/op   2284929 B/op   3983 allocs/op
Benchmark_Parsers_XMLParsing/rules_10-16     55   29089623 ns/op   2285676 B/op   4054 allocs/op
Benchmark_Parsers_XMLParsing/rules_100-16    50   26225184 ns/op   2297209 B/op   4775 allocs/op
Benchmark_Parsers_XMLParsing/rules_500-16    54   30522089 ns/op   2348453 B/op   7975 allocs/op
Benchmark_Parsers_XMLParsing/rules_1000-16   38   32811054 ns/op   2412510 B/op  11975 allocs/op
```
