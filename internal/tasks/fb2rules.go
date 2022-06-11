package tasks

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
	"github.com/egnd/go-xmlparse/fb2"
)

func SkipFB2Binaries(next xmlparse.TokenHandler) xmlparse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xmlparse.TokenReader) error {
		if _, ok := obj.(*fb2.File); ok && node.Name.Local == "binary" {
			return nil
		}

		return next(obj, node, r)
	}
}

func SkipFB2DocInfo(next xmlparse.TokenHandler) xmlparse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xmlparse.TokenReader) error {
		if _, ok := obj.(*fb2.Description); ok && node.Name.Local == "document-info" {
			return nil
		}

		return next(obj, node, r)
	}
}

func SkipFB2CustomInfo(next xmlparse.TokenHandler) xmlparse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xmlparse.TokenReader) error {
		if _, ok := obj.(*fb2.Description); ok && node.Name.Local == "custom-info" {
			return nil
		}

		return next(obj, node, r)
	}
}

func SkipFB2Cover(next xmlparse.TokenHandler) xmlparse.TokenHandler {
	return func(obj interface{}, node xml.StartElement, r xmlparse.TokenReader) error {
		if _, ok := obj.(*fb2.TitleInfo); ok && node.Name.Local == "coverpage" {
			return nil
		}

		return next(obj, node, r)
	}
}
