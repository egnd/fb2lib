package tasks

import (
	"encoding/xml"

	"github.com/egnd/go-fb2parse"
)

func SkipFB2Binaries(next fb2parse.TokenHandler) fb2parse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xml.TokenReader) error {
		if _, ok := obj.(*fb2parse.FB2File); ok && node.Name.Local == "binary" {
			return nil
		}

		return next(obj, node, r)
	}
}

func SkipFB2DocInfo(next fb2parse.TokenHandler) fb2parse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xml.TokenReader) error {
		if _, ok := obj.(*fb2parse.FB2Description); ok && node.Name.Local == "document-info" {
			return nil
		}

		return next(obj, node, r)
	}
}

func SkipFB2CustomInfo(next fb2parse.TokenHandler) fb2parse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xml.TokenReader) error {
		if _, ok := obj.(*fb2parse.FB2Description); ok && node.Name.Local == "custom-info" {
			return nil
		}

		return next(obj, node, r)
	}
}

func SkipFB2Cover(next fb2parse.TokenHandler) fb2parse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xml.TokenReader) error {
		if _, ok := obj.(*fb2parse.FB2TitleInfo); ok && node.Name.Local == "coverpage" {
			return nil
		}

		return next(obj, node, r)
	}
}
