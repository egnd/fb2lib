package fb2parse

import (
	"encoding/xml"
	"errors"
	"io"
)

// FB2Description struct of fb2 description.
// http://www.fictionbook.org/index.php/Элемент_description
type FB2Description struct {
	TitleInfo    []FB2TitleInfo  `xml:"title-info"`
	SrcTitleInfo []FB2TitleInfo  `xml:"src-title-info"`
	DocInfo      []FB2DocInfo    `xml:"document-info"`
	PublishInfo  []FB2Publisher  `xml:"publish-info"`
	CustomInfo   []FB2CustomInfo `xml:"custom-info"`
}

// NewFB2Description factory for FB2Description.
func NewFB2Description(
	tokenName string, reader xml.TokenReader, rules []HandlingRule,
) (res FB2Description, err error) {
	var token xml.Token

	handler := buildChain(rules, getFB2DescriptionHandler(rules))

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
			if typedToken.Name.Local == tokenName {
				break loop
			}
		}
	}

	return res, err
}

//nolint:forcetypeassert
func getFB2DescriptionHandler(rules []HandlingRule) TokenHandler { //nolint:cyclop
	var title FB2TitleInfo

	var doc FB2DocInfo

	var publish FB2Publisher

	var custom FB2CustomInfo

	return func(res interface{}, node xml.StartElement, reader xml.TokenReader) (err error) {
		switch node.Name.Local {
		case "title-info":
			if title, err = NewFB2TitleInfo(node.Name.Local, reader, rules); err == nil {
				res.(*FB2Description).TitleInfo = append(res.(*FB2Description).TitleInfo, title)
			}
		case "src-title-info":
			if title, err = NewFB2TitleInfo(node.Name.Local, reader, rules); err == nil {
				res.(*FB2Description).SrcTitleInfo = append(res.(*FB2Description).SrcTitleInfo, title)
			}
		case "document-info":
			if doc, err = NewFB2DocInfo(node.Name.Local, reader, rules); err == nil {
				res.(*FB2Description).DocInfo = append(res.(*FB2Description).DocInfo, doc)
			}
		case "publish-info":
			if publish, err = NewFB2Publisher(node.Name.Local, reader, rules); err == nil {
				res.(*FB2Description).PublishInfo = append(res.(*FB2Description).PublishInfo, publish)
			}
		case "custom-info":
			if custom, err = NewFB2CustomInfo(node, reader); err == nil {
				res.(*FB2Description).CustomInfo = append(res.(*FB2Description).CustomInfo, custom)
			}
		}

		return
	}
}
