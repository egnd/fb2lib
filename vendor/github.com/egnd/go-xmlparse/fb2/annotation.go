package fb2

import "github.com/egnd/go-xmlparse"

// Annotation struct of fb2 annotation.
// http://www.fictionbook.org/index.php/Элемент_annotation
type Annotation struct {
	HTML string `xml:",innerxml"`
}

// NewAnnotation factory for Annotation.
func NewAnnotation(tokenName string, reader xmlparse.TokenReader) (res Annotation, err error) {
	res.HTML, err = xmlparse.TokenRead(tokenName, reader)

	return
}
