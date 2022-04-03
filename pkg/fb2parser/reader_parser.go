package fb2parser

import (
	"encoding/xml"
	"io"
)

func FB2FromReader(data io.Reader) (*FB2File, error) {
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = CharsetReader

	var res FB2File
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
