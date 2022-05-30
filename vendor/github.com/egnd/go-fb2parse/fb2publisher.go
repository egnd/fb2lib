package fb2parse

import (
	"encoding/xml"
	"errors"
	"io"
)

// FB2Publisher struct of fb2 publisher info.
// http://www.fictionbook.org/index.php/Элемент_publish-info
type FB2Publisher struct {
	BookAuthor []string      `xml:"book-author"`
	BookName   []string      `xml:"book-name"`
	Publisher  []string      `xml:"publisher"`
	City       []string      `xml:"city"`
	Year       []string      `xml:"year"`
	ISBN       []string      `xml:"isbn"`
	Sequence   []FB2Sequence `xml:"sequence"`
}

// NewFB2Publisher factory for FB2Publisher.
func NewFB2Publisher(
	tokenName string, reader xml.TokenReader, rules []HandlingRule,
) (res FB2Publisher, err error) {
	var token xml.Token

	handler := buildChain(rules, getFB2PublisherHandler(rules))

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
			if err = handler(&res, typedToken, reader); err != nil {
				break loop
			}
		case xml.EndElement:
			// log.Println("----", typedToken.Name.Local)
			if typedToken.Name.Local == tokenName {
				break loop
			}
		}
	}

	return res, err
}

//nolint:forcetypeassert
func getFB2PublisherHandler(_ []HandlingRule) TokenHandler { //nolint:cyclop
	var seq FB2Sequence

	var strVal string

	return func(res interface{}, node xml.StartElement, reader xml.TokenReader) (err error) {
		switch node.Name.Local {
		case "book-author":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil {
				res.(*FB2Publisher).BookAuthor = append(res.(*FB2Publisher).BookAuthor, strVal)
			}
		case "book-name":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil {
				res.(*FB2Publisher).BookName = append(res.(*FB2Publisher).BookName, strVal)
			}
		case "publisher":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil {
				res.(*FB2Publisher).Publisher = append(res.(*FB2Publisher).Publisher, strVal)
			}
		case "city":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil {
				res.(*FB2Publisher).City = append(res.(*FB2Publisher).City, strVal)
			}
		case "year":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil {
				res.(*FB2Publisher).Year = append(res.(*FB2Publisher).Year, strVal)
			}
		case "isbn":
			if strVal, err = GetContent(node.Name.Local, reader); err == nil {
				res.(*FB2Publisher).ISBN = append(res.(*FB2Publisher).ISBN, strVal)
			}
		case "sequence":
			if seq, err = NewFB2Sequence(node); err == nil {
				res.(*FB2Publisher).Sequence = append(res.(*FB2Publisher).Sequence, seq)
			}
		}

		return
	}
}
