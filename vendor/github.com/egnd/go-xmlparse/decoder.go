// Package xmlparse contains tools for xml parsing
package xmlparse

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"golang.org/x/net/html/charset"
)

// NewDecoder creates decoder for xml data.
func NewDecoder(reader io.Reader) *xml.Decoder {
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity

	return decoder
}

// Unmarshal converts xml data to some struct.
func Unmarshal(data []byte, v any) (err error) {
	if err = NewDecoder(bytes.NewBuffer(data)).Decode(v); err != nil {
		err = fmt.Errorf("xmlparse unmarshal err: %w", err)
	}

	return
}
