// Package fb2parse contains tools for parsing fb2-files
package fb2parse

import (
	"encoding/xml"
	"errors"
	"io"
)

// FB2File struct of fb2 file.
// http://www.fictionbook.org/index.php/Элемент_FictionBook
// http://www.fictionbook.org/index.php/Описание_формата_FB2_от_Sclex
type FB2File struct {
	Description []FB2Description `xml:"description"`
	// Body        []FB2Body      `xml:"body"`
	Binary []FB2Binary `xml:"binary"`
}

// NewFB2File factory for FB2File.
func NewFB2File(doc *xml.Decoder, rules ...HandlingRule) (res FB2File, err error) {
	var token xml.Token

	handler := buildChain(rules, getFB2FileHandler(rules))

loop:
	for {
		if token, err = doc.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}

			break
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			if err = handler(&res, typedToken, doc); err != nil {
				break loop
			}
		}
	}

	return
}

//nolint:forcetypeassert
func getFB2FileHandler(rules []HandlingRule) TokenHandler {
	var binary FB2Binary

	var descr FB2Description

	return func(res interface{}, node xml.StartElement, reader xml.TokenReader) (err error) {
		switch node.Name.Local {
		case "description":
			if descr, err = NewFB2Description(node.Name.Local, reader, rules); err == nil {
				res.(*FB2File).Description = append(res.(*FB2File).Description, descr)
			}
		case "binary":
			if binary, err = NewFB2Binary(node, reader); err == nil {
				res.(*FB2File).Binary = append(res.(*FB2File).Binary, binary)
			}
		case "body":
			err = SkipToken(node.Name.Local, reader)
		}

		return
	}
}
