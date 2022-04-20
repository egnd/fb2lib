package entities

import (
	"github.com/egnd/fb2lib/pkg/fb2parser"
)

type FB2Book struct {
	Description struct {
		fb2parser.FB2Description
		TitleInfo    FB2TitleInfo  `xml:"title-info"`
		SrcTitleInfo *FB2TitleInfo `xml:"src-title-info"`
	} `xml:"description"`
	Binary []struct {
		ID   string `xml:"id,attr"`
		Type string `xml:"content-type,attr"`
		Data string `xml:",innerxml"`
	} `xml:"binary"`
}

type FB2TitleInfo struct {
	fb2parser.FB2TitleInfo
	Coverpage *FB2CoverPage `xml:"coverpage"`
}

type FB2CoverPage struct {
	Images []struct {
		Href string `xml:"href,attr"`
	} `xml:"image"`
}
