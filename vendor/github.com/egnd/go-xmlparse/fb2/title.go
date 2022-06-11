package fb2

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// TitleInfo struct of fb2 title info.
// http://www.fictionbook.org/index.php/Элемент_title-info
type TitleInfo struct {
	Genre      []string     `xml:"genre"`
	Author     []Author     `xml:"author"`
	BookTitle  []string     `xml:"book-title"`
	Annotation []Annotation `xml:"annotation"`
	Keywords   []string     `xml:"keywords"`
	Date       []string     `xml:"date"`
	Coverpage  []Cover      `xml:"coverpage"`
	Lang       []string     `xml:"lang"`
	SrcLang    []string     `xml:"src-lang"`
	Translator []Author     `xml:"translator"`
	Sequence   []Sequence   `xml:"sequence"`
}

// NewTitleInfo factory for TitleInfo.
func NewTitleInfo(
	tokenName string, reader xmlparse.TokenReader, rules []xmlparse.Rule,
) (res TitleInfo, err error) {
	var token xml.Token

	handler := xmlparse.WrapRules(rules, getTitleInfoHandler(rules))

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
func getTitleInfoHandler(_ []xmlparse.Rule) xmlparse.TokenHandler { //nolint:cyclop,gocognit
	var strVal string

	var author Author

	var seq Sequence

	var annotation Annotation

	var cover Cover

	return func(res interface{}, node xml.StartElement, reader xmlparse.TokenReader) (err error) {
		switch node.Name.Local {
		case "genre":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*TitleInfo).Genre = append(res.(*TitleInfo).Genre, strVal)
			}
		case "author":
			if author, err = NewAuthor(node.Name.Local, reader); err == nil {
				res.(*TitleInfo).Author = append(res.(*TitleInfo).Author, author)
			}
		case "book-title":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*TitleInfo).BookTitle = append(res.(*TitleInfo).BookTitle, strVal)
			}
		case "annotation":
			if annotation, err = NewAnnotation(node.Name.Local, reader); err == nil {
				res.(*TitleInfo).Annotation = append(res.(*TitleInfo).Annotation, annotation)
			}
		case "keywords":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*TitleInfo).Keywords = append(res.(*TitleInfo).Keywords, strVal)
			}
		case "date":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*TitleInfo).Date = append(res.(*TitleInfo).Date, strVal)
			}
		case "coverpage":
			if cover, err = NewCover(node.Name.Local, reader); err == nil {
				res.(*TitleInfo).Coverpage = append(res.(*TitleInfo).Coverpage, cover)
			}
		case "lang":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*TitleInfo).Lang = append(res.(*TitleInfo).Lang, strVal)
			}
		case "src-lang":
			if strVal, err = xmlparse.TokenRead(node.Name.Local, reader); err == nil && strVal != "" {
				res.(*TitleInfo).SrcLang = append(res.(*TitleInfo).SrcLang, strVal)
			}
		case "translator":
			if author, err = NewAuthor(node.Name.Local, reader); err == nil {
				res.(*TitleInfo).Translator = append(res.(*TitleInfo).Translator, author)
			}
		case "sequence":
			if seq, err = NewSequence(node); err == nil {
				res.(*TitleInfo).Sequence = append(res.(*TitleInfo).Sequence, seq)
			}
		}

		return
	}
}
