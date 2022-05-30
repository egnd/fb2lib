package fb2parse

import (
	"encoding/xml"
	"errors"
	"io"
)

// FB2DocInfo struct of fb2 document info.
// http://www.fictionbook.org/index.php/Элемент_document-info
type FB2DocInfo struct {
	Authors    []FB2Author `xml:"author"`
	SrcURL     []string    `xml:"src-url"`
	ID         []string    `xml:"id"`
	Version    []string    `xml:"version"`
	Publishers []FB2Author `xml:"publisher"`
	// program-used - 0..1 (один, опционально) @TODO:
	// date - 1 (один, обязателен) @TODO:
	// src-ocr - 0..1 (один, опционально) @TODO:
	// history - 0..1 (один, опционально) @TODO:
}

// NewFB2DocInfo factory for NewFB2DocInfo.
func NewFB2DocInfo(
	tokenName string, reader xml.TokenReader, rules []HandlingRule,
) (res FB2DocInfo, err error) {
	var token xml.Token

	handler := buildChain(rules, getFB2DocInfoHandler(rules))

loop:
	for {
		if token, err = reader.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}

			break
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			if err = handler(&res, typedToken, reader); err != nil {
				break loop
			}
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				break loop
			}
		}
	}

	return res, err
}

//nolint:forcetypeassert
func getFB2DocInfoHandler(_ []HandlingRule) TokenHandler { //nolint:cyclop
	var strVal string

	var author FB2Author

	return func(res interface{}, node xml.StartElement, reader xml.TokenReader) (err error) {
		switch node.Name.Local {
		case "author":
			if author, err = NewFB2Author(node.Name.Local, reader); err == nil {
				res.(*FB2DocInfo).Authors = append(res.(*FB2DocInfo).Authors, author)
			}
		case "src-url":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*FB2DocInfo).SrcURL = append(res.(*FB2DocInfo).SrcURL, strVal)
			}
		case "id":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*FB2DocInfo).ID = append(res.(*FB2DocInfo).ID, strVal)
			}
		case "version":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*FB2DocInfo).Version = append(res.(*FB2DocInfo).Version, strVal)
			}
		case "publisher":
			if author, err = NewFB2Author(node.Name.Local, reader); err == nil {
				res.(*FB2DocInfo).Publishers = append(res.(*FB2DocInfo).Publishers, author)
			}
		}

		return
	}
}
