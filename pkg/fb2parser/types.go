package fb2parser

import (
	"encoding/xml"
	"io"
)

// http://www.fictionbook.org/index.php/%D0%9E%D0%BF%D0%B8%D1%81%D0%B0%D0%BD%D0%B8%D0%B5_%D1%84%D0%BE%D1%80%D0%BC%D0%B0%D1%82%D0%B0_FB2_%D0%BE%D1%82_Sclex
type FB2File struct {
	Description FB2Description `xml:"description"`
}

type FB2Description struct {
	TitleInfo    FB2TitleInfo  `xml:"title-info"`
	SrcTitleInfo *FB2TitleInfo `xml:"src-title-info"`
	PublishInfo  *FB2Publisher `xml:"publish-info"`
}

func ParseFB2Description(curTag string, decoder xml.TokenReader) (res FB2Description, err error) {
	for {
		var token xml.Token
		if token, err = decoder.Token(); err != nil {
			if err == io.EOF {
				err = nil
			}

			return
		}

		switch tokType := token.(type) {
		case xml.StartElement:
			switch tokType.Name.Local {
			case "title-info":
				if res.TitleInfo, err = ParseFB2TitleInfo(tokType.Name.Local, decoder); err != nil {
					return
				}
			case "src-title-info":
				var item FB2TitleInfo
				if item, err = ParseFB2TitleInfo(tokType.Name.Local, decoder); err != nil {
					return
				}

				res.SrcTitleInfo = &item
			case "publish-info":
				var item FB2Publisher
				if item, err = ParseFB2Publisher(tokType.Name.Local, decoder); err != nil {
					return
				}

				res.PublishInfo = &item
			}
		case xml.EndElement:
			if tokType.Name.Local == curTag {
				return
			}
		}
	}
}

type FB2TitleInfo struct {
	BookTitle  string        `xml:"book-title"`
	Keywords   string        `xml:"keywords"`
	Date       string        `xml:"date"`
	Lang       string        `xml:"lang"`
	SrcLang    string        `xml:"src-lang"`
	Genre      []string      `xml:"genre"`
	Author     []FB2Author   `xml:"author"`
	Translator []FB2Author   `xml:"translator"`
	Sequence   []FB2Sequence `xml:"sequence"`
	Annotation struct {
		HTML string `xml:",innerxml"`
	} `xml:"annotation"`
}

func ParseFB2TitleInfo(curTag string, reader xml.TokenReader) (res FB2TitleInfo, err error) {
	for {
		var token xml.Token
		if token, err = reader.Token(); err != nil {
			if err == io.EOF {
				err = nil
			}

			return
		}

		switch tokType := token.(type) {
		case xml.StartElement:
			switch tokType.Name.Local {
			case "genre":
				if val := GetTokenValue(tokType.Name.Local, reader); val != "" {
					res.Genre = append(res.Genre, val)
				}
			case "book-title":
				res.BookTitle = GetTokenValue(tokType.Name.Local, reader)
			case "date":
				res.Date = GetTokenValue(tokType.Name.Local, reader)
			case "lang":
				res.Lang = GetTokenValue(tokType.Name.Local, reader)
			case "src-lang":
				res.SrcLang = GetTokenValue(tokType.Name.Local, reader)
			case "annotation":
				res.Annotation.HTML = GetTokenValue(tokType.Name.Local, reader)
			case "author":
				var item FB2Author
				if item, err = ParseFB2Author(tokType.Name.Local, reader); err != nil {
					return
				}

				res.Author = append(res.Translator, item)
			case "translator":
				var item FB2Author
				if item, err = ParseFB2Author(tokType.Name.Local, reader); err != nil {
					return
				}

				res.Translator = append(res.Translator, item)
			case "sequence":
				res.Sequence = append(res.Sequence, ParseFB2Sequence(tokType))
			}
		case xml.EndElement:
			if tokType.Name.Local == curTag {
				return
			}
		}
	}
}

type FB2Author struct {
	ID         string `xml:"id"`
	FirstName  string `xml:"first-name"`
	MiddleName string `xml:"middle-name"`
	LastName   string `xml:"last-name"`
	NickName   string `xml:"nickname"`
}

func ParseFB2Author(curTag string, reader xml.TokenReader) (res FB2Author, err error) {
	for {
		var token xml.Token
		if token, err = reader.Token(); err != nil {
			if err == io.EOF {
				err = nil
			}

			return
		}

		switch tokType := token.(type) {
		case xml.StartElement:
			switch tokType.Name.Local {
			case "id":
				res.ID = GetTokenValue(tokType.Name.Local, reader)
			case "first-name":
				res.FirstName = GetTokenValue(tokType.Name.Local, reader)
			case "middle-name":
				res.MiddleName = GetTokenValue(tokType.Name.Local, reader)
			case "last-name":
				res.LastName = GetTokenValue(tokType.Name.Local, reader)
			case "nickname":
				res.NickName = GetTokenValue(tokType.Name.Local, reader)
			}
		case xml.EndElement:
			if tokType.Name.Local == curTag {
				return
			}
		}
	}
}

type FB2Sequence struct {
	Number string `xml:"number,attr"`
	Name   string `xml:"name,attr"`
}

func ParseFB2Sequence(token xml.StartElement) (res FB2Sequence) {
	for _, attr := range token.Attr {
		switch attr.Name.Local {
		case "name":
			res.Name = attr.Value
		case "number":
			res.Number = attr.Value
		}
	}

	return
}

type FB2Publisher struct {
	BookName  string `xml:"book-name"`
	Publisher string `xml:"publisher"`
	City      string `xml:"city"`
	Year      string `xml:"year"`
	ISBN      string `xml:"isbn"`
}

func ParseFB2Publisher(curTag string, reader xml.TokenReader) (res FB2Publisher, err error) {
	for {
		var token xml.Token
		if token, err = reader.Token(); err != nil {
			if err == io.EOF {
				err = nil
			}

			return
		}

		switch tokType := token.(type) {
		case xml.StartElement:
			switch tokType.Name.Local {
			case "book-name":
				res.BookName = GetTokenValue(tokType.Name.Local, reader)
			case "publisher":
				res.Publisher = GetTokenValue(tokType.Name.Local, reader)
			case "city":
				res.City = GetTokenValue(tokType.Name.Local, reader)
			case "year":
				res.Year = GetTokenValue(tokType.Name.Local, reader)
			case "isbn":
				res.ISBN = GetTokenValue(tokType.Name.Local, reader)
			}
		case xml.EndElement:
			if tokType.Name.Local == curTag {
				return
			}
		}
	}
}
