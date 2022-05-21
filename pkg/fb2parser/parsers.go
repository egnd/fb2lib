package fb2parser

import (
	"encoding/xml"
	"io"

	"golang.org/x/net/html/charset"
)

func UnmarshalStream(data io.Reader, v interface{}) error {
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity

	if err := decoder.Decode(v); err != nil {
		return err
	}

	return nil
}
