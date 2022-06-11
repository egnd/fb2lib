package fb2

import (
	"encoding/xml"
)

// Image struct of fb2 image.
// http://www.fictionbook.org/index.php/Элемент_image
type Image struct {
	Type  string `xml:"type,attr"`
	Href  string `xml:"href,attr"`
	Alt   string `xml:"alt,attr"`
	Title string `xml:"title,attr"`
	ID    string `xml:"id,attr"`
}

// NewImage factory for Image.
func NewImage(token xml.StartElement) (res Image) {
	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case "type":
			res.Type = attr.Value
		case "href":
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
