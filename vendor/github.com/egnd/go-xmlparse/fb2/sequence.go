package fb2

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// Sequence struct of fb2 sequence info.
// http://www.fictionbook.org/index.php/Элемент_sequence
type Sequence struct {
	Number string `xml:"number,attr"`
	Name   string `xml:"name,attr"`
}

func (s Sequence) String() string {
	var buf bytes.Buffer

	if s.Name == "" {
		return ""
	}

	var num string
	if s.Number != "" && s.Number != "0" {
		num = fmt.Sprintf("(%s)", s.Number)
	}

	for _, subitem := range strings.Split(s.Name, ",") {
		if subitem = strings.TrimSpace(subitem); subitem == "" {
			continue
		}

		buf.WriteString(subitem)

		if num != "" {
			buf.WriteRune(' ')
			buf.WriteString(num)
		}

		return buf.String()
	}

	return ""
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
