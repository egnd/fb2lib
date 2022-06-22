package fb2

import (
	"bytes"
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// Publisher struct of fb2 publisher info.
// http://www.fictionbook.org/index.php/Элемент_publish-info
type Publisher struct {
	BookAuthor []string   `xml:"book-author"`
	BookName   []string   `xml:"book-name"`
	Publisher  []string   `xml:"publisher"`
	City       []string   `xml:"city"`
	Year       []string   `xml:"year"`
	ISBN       []string   `xml:"isbn"`
	Sequence   []Sequence `xml:"sequence"`
}

func (p Publisher) String() string {
	var buf bytes.Buffer

	var strVal string

	if strVal = xmlparse.GetStrFrom(p.Publisher); strVal != "" {
		buf.WriteString(strVal)
	}

	if strVal = xmlparse.GetStrFrom(p.City); strVal != "" {
		if buf.Len() > 0 {
			buf.WriteString(" (")
			buf.WriteString(strVal)
			buf.WriteRune(')')
		} else {
			buf.WriteString(strVal)
		}
	}

	return buf.String()
}

// NewPublisher factory for Publisher.
func NewPublisher(
	tokenName string, reader xmlparse.TokenReader, rules []xmlparse.Rule,
) (res Publisher, err error) {
	var token xml.Token

	handler := xmlparse.WrapRules(rules, getPublisherHandler(rules))

	for {
		if token, err = reader.Token(); err != nil {
			return
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			if err = handler(&res, typedToken, reader); err != nil {
				return
			}
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				return
			}
		}
	}
}

//nolint:forcetypeassert
func getPublisherHandler(_ []xmlparse.Rule) xmlparse.TokenHandler { //nolint:cyclop
	var seq Sequence

	var strVal string

	return func(res interface{}, node xml.StartElement, reader xmlparse.TokenReader) (err error) {
		switch node.Name.Local {
		case "book-author":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil {
				res.(*Publisher).BookAuthor = append(res.(*Publisher).BookAuthor, strVal)
			}
		case "book-name":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil {
				res.(*Publisher).BookName = append(res.(*Publisher).BookName, strVal)
			}
		case "publisher":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil {
				res.(*Publisher).Publisher = append(res.(*Publisher).Publisher, strVal)
			}
		case "city":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil {
				res.(*Publisher).City = append(res.(*Publisher).City, strVal)
			}
		case "year":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil {
				res.(*Publisher).Year = append(res.(*Publisher).Year, strVal)
			}
		case "isbn":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil {
				res.(*Publisher).ISBN = append(res.(*Publisher).ISBN, strVal)
			}
		case "sequence":
			if seq, err = NewSequence(node); err == nil {
				res.(*Publisher).Sequence = append(res.(*Publisher).Sequence, seq)
			}
		}

		return
	}
}
