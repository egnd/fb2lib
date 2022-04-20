package fb2parser

// http://www.fictionbook.org/index.php/%D0%9E%D0%BF%D0%B8%D1%81%D0%B0%D0%BD%D0%B8%D0%B5_%D1%84%D0%BE%D1%80%D0%BC%D0%B0%D1%82%D0%B0_FB2_%D0%BE%D1%82_Sclex
type FB2File struct {
	Description FB2Description `xml:"description"`
}

type FB2Description struct {
	TitleInfo    FB2TitleInfo  `xml:"title-info"`
	SrcTitleInfo *FB2TitleInfo `xml:"src-title-info"`
	PublishInfo  *FB2Publisher `xml:"publish-info"`
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

type FB2Author struct {
	FirstName  string `xml:"first-name"`
	MiddleName string `xml:"middle-name"`
	LastName   string `xml:"last-name"`
}

type FB2Sequence struct {
	Number string `xml:"number,attr"`
	Name   string `xml:"name,attr"`
}

type FB2Publisher struct {
	BookName  string `xml:"book-name"`
	Publisher string `xml:"publisher"`
	City      string `xml:"city"`
	Year      string `xml:"year"`
	ISBN      string `xml:"isbn"`
}
