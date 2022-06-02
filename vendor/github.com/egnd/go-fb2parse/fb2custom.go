package fb2parse

import "encoding/xml"

// FB2CustomInfo struct of fb2 custom info.
// http://www.fictionbook.org/index.php/Элемент_custom-info
type FB2CustomInfo struct {
	InfoType string `xml:"info-type,attr"`
	Data     string `xml:",innerxml"`
}

// NewFB2CustomInfo factory for FB2CustomInfo.
func NewFB2CustomInfo(token xml.StartElement, reader xml.TokenReader) (res FB2CustomInfo, err error) {
	if res.Data, err = GetContent(token.Name.Local, reader); err != nil {
		return
	}

	for _, attr := range token.Attr {
		if attr.Name.Local == "info-type" {
			res.InfoType = attr.Value

			break
		}
	}

	return
}
