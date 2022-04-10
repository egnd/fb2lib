package fb2parser

import (
	"encoding/xml"
	"io"

	"golang.org/x/net/html/charset"
)

func UnmarshalFB2Stream(data io.Reader) (*FB2File, error) {
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel
	// decoder.CharsetReader = CharsetReader

	var res FB2File
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

func ParseFB2Stream(data io.Reader) (res *FB2File, err error) {
	res = &FB2File{}
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = CharsetReader

	for {
		var token xml.Token
		if token, err = decoder.Token(); err != nil {
			if err == io.EOF {
				err = nil
			}

			return
		}

		switch tokType := token.(type) {
		case xml.StartElement:
			if tokType.Name.Local == "description" {
				res.Description, err = ParseFB2Description(tokType.Name.Local, decoder)
				if err != nil {
					return
				}
			}
		case xml.EndElement:
			if tokType.Name.Local == "description" {
				return
			}
		}
	}
}
