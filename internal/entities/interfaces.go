package entities

import "github.com/blevesearch/bleve/v2"

type IIndexFactory func(string) (bleve.Index, error)
