package fb2parse

import (
	"encoding/xml"
)

// FB2Annotation struct of fb2 annotation.
// http://www.fictionbook.org/index.php/Элемент_annotation
type FB2Annotation struct {
	HTML string `xml:",innerxml"`
}

// NewFB2Annotation factory for FB2Annotation.
func NewFB2Annotation(tokenName string, reader xml.TokenReader) (res FB2Annotation, err error) {
	res.HTML, err = GetContent(tokenName, reader)

	return
}
