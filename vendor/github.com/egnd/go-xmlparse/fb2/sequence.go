package fb2

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// Sequence struct of fb2 sequence info.
// http://www.fictionbook.org/index.php/Элемент_sequence
type Sequence struct {
	Number string `xml:"number,attr"`
	Name   string `xml:"name,attr"`
}

// NewSequence factory for Sequence.
func NewSequence(token xml.StartElement) (res Sequence, err error) {
	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case "name":
			res.Name = attr.Value
		case "number":
			if num, err := strconv.Atoi(attr.Value); err == nil && num > 0 {
				res.Number = fmt.Sprint(num)
			}
		}
	}

	return
}
