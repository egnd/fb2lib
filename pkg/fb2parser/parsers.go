package fb2parser

import (
	"encoding/xml"
	"io"

	"golang.org/x/net/html/charset"
)

func UnmarshalStream(data io.Reader) (*FB2File, error) {
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel

	var res FB2File
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
