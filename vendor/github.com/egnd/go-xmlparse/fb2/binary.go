package fb2

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// Binary struct of fb2 binary data.
// http://www.fictionbook.org/index.php/Элемент_binary
type Binary struct {
	ContentType string `xml:"content-type,attr"`
	ID          string `xml:"id,attr"`
	Data        string `xml:",innerxml"`
}

// NewBinary factory for Binary.
func NewBinary(token xml.StartElement, reader xmlparse.TokenReader) (res Binary, err error) {
	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case "content-type":
			res.ContentType = attr.Value
		case "id":
			res.ID = attr.Value
		}
	}

	res.Data, err = xmlparse.TokenRead(token.Name.Local, reader)

	return
}
