package xmlparse

import (
	"encoding/xml"
	"strings"
)

// TokenHandler handler for fb2 tokens.
type TokenHandler func(interface{}, xml.StartElement, TokenReader) error

// TokenReader token reader interface.
type TokenReader interface {
	xml.TokenReader
}

// TokenSkip skips current token at reader.
func TokenSkip(tokenName string, reader TokenReader) (err error) {
	var token xml.Token

	for {
		if token, err = reader.Token(); err != nil {
			break
		}

		if elem, ok := token.(xml.EndElement); ok && elem.Name.Local == tokenName {
			break
		}
	}

	return
}

// TokenRead returns xml token inner content.
func TokenRead(tokenName string, reader TokenReader) (res string, err error) {
	var token xml.Token

	var buf strings.Builder

	for {
		if token, err = reader.Token(); err != nil {
			return
		}

		switch typedToken := token.(type) {
		case xml.CharData:
			buf.Write(typedToken)
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				return strings.TrimSpace(buf.String()), nil
			}
		}
	}
}
