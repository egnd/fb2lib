package library

import (
	"encoding/xml"
	"io"

	"gitlab.com/egnd/bookshelf/internal/entities"
)

func NewFB2FileFromReader(data io.Reader) (*entities.FB2File, error) {
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = CharsetReader

	var res entities.FB2File
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}
