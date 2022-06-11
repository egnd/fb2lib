package fb2

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// Description struct of fb2 description.
// http://www.fictionbook.org/index.php/Элемент_description
type Description struct {
	TitleInfo    []TitleInfo  `xml:"title-info"`
	SrcTitleInfo []TitleInfo  `xml:"src-title-info"`
	DocInfo      []DocInfo    `xml:"document-info"`
	PublishInfo  []Publisher  `xml:"publish-info"`
	CustomInfo   []CustomInfo `xml:"custom-info"`
}

// NewDescription factory for Description.
func NewDescription(
	tokenName string, reader xmlparse.TokenReader, rules []xmlparse.Rule,
) (res Description, err error) {
	var token xml.Token

	handler := xmlparse.WrapRules(rules, getDescriptionHandler(rules))

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
func getDescriptionHandler(rules []xmlparse.Rule) xmlparse.TokenHandler { //nolint:cyclop
	var title TitleInfo

	var doc DocInfo

	var publish Publisher

	var custom CustomInfo

	return func(res interface{}, node xml.StartElement, reader xmlparse.TokenReader) (err error) {
		switch node.Name.Local {
		case "title-info":
			if title, err = NewTitleInfo(node.Name.Local, reader, rules); err == nil {
				res.(*Description).TitleInfo = append(res.(*Description).TitleInfo, title)
			}
		case "src-title-info":
			if title, err = NewTitleInfo(node.Name.Local, reader, rules); err == nil {
				res.(*Description).SrcTitleInfo = append(res.(*Description).SrcTitleInfo, title)
			}
		case "document-info":
			if doc, err = NewDocInfo(node.Name.Local, reader, rules); err == nil {
				res.(*Description).DocInfo = append(res.(*Description).DocInfo, doc)
			}
		case "publish-info":
			if publish, err = NewPublisher(node.Name.Local, reader, rules); err == nil {
				res.(*Description).PublishInfo = append(res.(*Description).PublishInfo, publish)
			}
		case "custom-info":
			if custom, err = NewCustomInfo(node, reader); err == nil {
				res.(*Description).CustomInfo = append(res.(*Description).CustomInfo, custom)
			}
		}

		return
	}
}
