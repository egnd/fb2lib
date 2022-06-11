package fb2

import "github.com/egnd/go-xmlparse"

// Body struct of fb2 body.
// http://www.fictionbook.org/index.php/Элемент_body
type Body struct {
	HTML string `xml:",innerxml"`
}

// NewBody factory for Body.
func NewBody(tokenName string, reader xmlparse.TokenReader) (res Body, err error) {
	res.HTML, err = xmlparse.TokenRead(tokenName, reader)

	return
}
