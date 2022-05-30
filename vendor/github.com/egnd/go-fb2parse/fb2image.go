package fb2parse

import (
	"encoding/xml"
)

// FB2Image struct of fb2 image.
// http://www.fictionbook.org/index.php/Элемент_image
type FB2Image struct {
	Type  string `xml:"l:type,attr"`
	Href  string `xml:"l:href,attr"`
	Alt   string `xml:"alt,attr"`
	Title string `xml:"title,attr"`
	ID    string `xml:"id,attr"`
}

// NewFB2Image factory for FB2Image.
func NewFB2Image(token xml.StartElement) (res FB2Image, err error) {
	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case "l:type":
			res.Type = attr.Value
		case "l:href":
			res.Href = attr.Value
		case "alt":
			res.Alt = attr.Value
		case "title":
			res.Title = attr.Value
		case "id":
			res.ID = attr.Value
		}
	}

	return
}
