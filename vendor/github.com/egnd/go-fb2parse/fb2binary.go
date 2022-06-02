package fb2parse

import "encoding/xml"

// FB2Binary struct of fb2 binary data.
// http://www.fictionbook.org/index.php/Элемент_binary
type FB2Binary struct {
	ContentType string `xml:"content-type,attr"`
	ID          string `xml:"id,attr"`
	Data        string `xml:",innerxml"`
}

// NewFB2Binary factory for FB2Binary.
func NewFB2Binary(token xml.StartElement, reader xml.TokenReader) (res FB2Binary, err error) {
	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case "content-type":
			res.ContentType = attr.Value
		case "id":
			res.ID = attr.Value
		}
	}

	res.Data, err = GetContent(token.Name.Local, reader)

	return
}
