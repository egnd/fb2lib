package fb2

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// DocInfo struct of fb2 document info.
// http://www.fictionbook.org/index.php/Элемент_document-info
type DocInfo struct {
	Authors    []Author `xml:"author"`
	SrcURL     []string `xml:"src-url"`
	ID         []string `xml:"id"`
	Version    []string `xml:"version"`
	Publishers []Author `xml:"publisher"`
	// program-used - 0..1 (один, опционально) @TODO:
	// date - 1 (один, обязателен) @TODO:
	// src-ocr - 0..1 (один, опционально) @TODO:
	// history - 0..1 (один, опционально) @TODO:
}

// NewDocInfo factory for NewDocInfo.
func NewDocInfo(
	tokenName string, reader xmlparse.TokenReader, rules []xmlparse.Rule,
) (res DocInfo, err error) {
	var token xml.Token

	handler := xmlparse.WrapRules(rules, getDocInfoHandler(rules))

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
func getDocInfoHandler(_ []xmlparse.Rule) xmlparse.TokenHandler { //nolint:cyclop
	var strVal string

	var author Author

	return func(res interface{}, node xml.StartElement, reader xmlparse.TokenReader) (err error) {
		switch node.Name.Local {
		case "author":
			if author, err = NewAuthor(node.Name.Local, reader); err == nil {
				res.(*DocInfo).Authors = append(res.(*DocInfo).Authors, author)
			}
		case "src-url":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*DocInfo).SrcURL = append(res.(*DocInfo).SrcURL, strVal)
			}
		case "id":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*DocInfo).ID = append(res.(*DocInfo).ID, strVal)
			}
		case "version":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*DocInfo).Version = append(res.(*DocInfo).Version, strVal)
			}
		case "publisher":
			if author, err = NewAuthor(node.Name.Local, reader); err == nil {
				res.(*DocInfo).Publishers = append(res.(*DocInfo).Publishers, author)
			}
		}

		return
	}
}
