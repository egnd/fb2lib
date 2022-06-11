// Package fb2 contains entities for fb2-files parsing
package fb2

import (
	"encoding/xml"
	"errors"
	"io"

	"github.com/egnd/go-xmlparse"
)

// File struct of fb2 file.
// http://www.fictionbook.org/index.php/Элемент_FictionBook
// http://www.fictionbook.org/index.php/Описание_формата__от_Sclex
type File struct {
	Description []Description `xml:"description"`
	// Body        []Body      `xml:"body"`
	Binary []Binary `xml:"binary"`
}

// NewFile factory for File.
func NewFile(reader xmlparse.TokenReader, rules ...xmlparse.Rule) (res File, err error) {
	var token xml.Token

	handler := xmlparse.WrapRules(rules, getFileHandler(rules))

	for {
		if token, err = reader.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}

			return
		}

		if typedToken, ok := token.(xml.StartElement); ok {
			if err = handler(&res, typedToken, reader); err != nil {
				return
			}
		}
	}
}

//nolint:forcetypeassert
func getFileHandler(rules []xmlparse.Rule) xmlparse.TokenHandler {
	var binary Binary

	var descr Description

	return func(res interface{}, node xml.StartElement, reader xmlparse.TokenReader) (err error) {
		switch node.Name.Local {
		case "description":
			if descr, err = NewDescription(node.Name.Local, reader, rules); err == nil {
				res.(*File).Description = append(res.(*File).Description, descr)
			}
		case "binary":
			if binary, err = NewBinary(node, reader); err == nil {
				res.(*File).Binary = append(res.(*File).Binary, binary)
			}
		case "body":
			err = xmlparse.TokenSkip(node.Name.Local, reader)
		}

		return
	}
}
