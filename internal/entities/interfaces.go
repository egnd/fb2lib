package entities

import "github.com/blevesearch/bleve/v2"

type IBleveIndex interface {
	bleve.Index
}

type IBooksRepo interface{}
