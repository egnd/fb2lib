package fb2

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// Cover struct of fb2 cover page.
// http://www.fictionbook.org/index.php/Элемент_coverpage
type Cover struct {
	Images []Image `xml:"image"`
}

// NewCover factory for Cover.
func NewCover(tokenName string, reader xmlparse.TokenReader) (res Cover, err error) {
	var token xml.Token

	for {
		if token, err = reader.Token(); err != nil {
			return
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			switch typedToken.Name.Local { //nolint:gocritic
			case "image":
				res.Images = append(res.Images, NewImage(typedToken))
			}
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				return res, nil
			}
		}
	}
}
