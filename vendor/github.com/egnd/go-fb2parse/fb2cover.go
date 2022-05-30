package fb2parse

import (
	"encoding/xml"
	"errors"
	"io"
)

// FB2Cover struct of fb2 cover page.
// http://www.fictionbook.org/index.php/Элемент_coverpage
type FB2Cover struct {
	Images []FB2Image `xml:"image"`
}

// NewFB2Cover factory for FB2Cover.
func NewFB2Cover(tokenName string, reader xml.TokenReader) (res FB2Cover, err error) {
	var token xml.Token

	var image FB2Image

loop:
	for {
		if token, err = reader.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}

			break
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			switch typedToken.Name.Local { //nolint:gocritic
			case "image":
				if image, err = NewFB2Image(typedToken); err == nil {
					res.Images = append(res.Images, image)
				}
			}

			if err != nil {
				break loop
			}
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				break loop
			}
		}
	}

	return res, err
}
