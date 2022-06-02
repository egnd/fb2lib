package fb2parse

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"

	"golang.org/x/net/html/charset"
)

func buildChain(rules []HandlingRule, handler TokenHandler) TokenHandler {
	for i := len(rules) - 1; i >= 0; i-- {
		handler = rules[i](handler)
	}

	return handler
}

// NewDecoder creates decoder for fb2 xml data.
func NewDecoder(reader io.Reader) *xml.Decoder {
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity

	return decoder
}

// Unmarshal converts fb2 xml data some struct.
func Unmarshal(data []byte, v any) error {
	return NewDecoder(bytes.NewBuffer(data)).Decode(v) //nolint:wrapcheck
}

// GetContent returns xml token inner content.
func GetContent(tokenName string, reader xml.TokenReader) (res string, err error) {
	var token xml.Token

	var buf strings.Builder

loop:
	for {
		if token, err = reader.Token(); err != nil {
			break
		}

		switch typedToken := token.(type) {
		case xml.CharData:
			buf.Write(typedToken)
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				break loop
			}
		}
	}

	return strings.TrimSpace(buf.String()), err
}

// SkipToken skips current token at reader.
func SkipToken(tokenName string, reader xml.TokenReader) (err error) {
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
