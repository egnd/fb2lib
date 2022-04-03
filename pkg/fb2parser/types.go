package fb2parser

// http://www.fictionbook.org/index.php/%D0%9E%D0%BF%D0%B8%D1%81%D0%B0%D0%BD%D0%B8%D0%B5_%D1%84%D0%BE%D1%80%D0%BC%D0%B0%D1%82%D0%B0_FB2_%D0%BE%D1%82_Sclex
type FB2File struct {
	Description FB2Description `xml:"description"`
	Binary      []FB2Binary    `xml:"binary"`
}

type FB2Binary struct {
	ID          string `xml:"id,attr"`
	ContentType string `xml:"content-type,attr"`
	Data        string `xml:",innerxml"`
}

type FB2Description struct {
	TitleInfo    FB2TitleInfo  `xml:"title-info"`
	SrcTitleInfo *FB2TitleInfo `xml:"src-title-info"`
	DocInfo      FB2DocInfo    `xml:"document-info"`
	PublishInfo  *FB2Publisher `xml:"publish-info"`
	CustomInfo   []struct {
		Data string `xml:",innerxml"`
	} `xml:"custom-info"`
}

type FB2TitleInfo struct {
	Genre      []string    `xml:"genre"`
	Author     []FB2Author `xml:"author"`
	BookTitle  string      `xml:"book-title"`
	Annotation struct {
		HTML string `xml:",innerxml"`
	} `xml:"annotation"`
	Keywords  string `xml:"keywords"`
	Date      string `xml:"date"`
	Coverpage struct {
		Image struct {
			Href string `xml:"href,attr"`
		} `xml:"image"`
	} `xml:"coverpage"`
	Lang       string        `xml:"lang"`
	SrcLang    string        `xml:"src-lang"`
	Translator []FB2Author   `xml:"translator"`
	Sequence   []FB2Sequence `xml:"sequence"`
}

type FB2Author struct {
	ID         string   `xml:"id"`
	FirstName  string   `xml:"first-name"`
	MiddleName string   `xml:"middle-name"`
	LastName   string   `xml:"last-name"`
	NickName   string   `xml:"nickname"`
	HomePage   []string `xml:"home-page"`
	Email      []string `xml:"email"`
}

type FB2Sequence struct {
	Number int    `xml:"number,attr"`
	Name   string `xml:"name,attr"`
}

type FB2DocInfo struct {
	Author      []FB2Author `xml:"author"`
	ProgramUsed string      `xml:"program-used"`
	Date        string      `xml:"date"`
	SrcURL      []string    `xml:"src-url"`
	SrcOCR      string      `xml:"src-ocr"`
	ID          string      `xml:"id"`
	Version     string      `xml:"version"`
	History     struct {
		HTML string `xml:",innerxml"`
	} `xml:"history"`
	Publisher []FB2Publisher `xml:"publisher"`
}

type FB2Publisher struct {
	BookName  string `xml:"book-name"`
	Publisher string `xml:"publisher"`
	City      string `xml:"city"`
	Year      int    `xml:"year"`
	ISBN      string `xml:"isbn"`
}
