package fb2parse

import (
	"encoding/xml"
)

// FB2Body struct of fb2 body.
// http://www.fictionbook.org/index.php/Элемент_body
type FB2Body struct {
	HTML string `xml:",innerxml"`
}

// NewFB2Body factory for FB2Body.
func NewFB2Body(tokenName string, reader xml.TokenReader) (res FB2Body, err error) {
	res.HTML, err = GetContent(tokenName, reader)

	return
}
