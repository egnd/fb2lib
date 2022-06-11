package fb2

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// CustomInfo struct of fb2 custom info.
// http://www.fictionbook.org/index.php/Элемент_custom-info
type CustomInfo struct {
	InfoType string `xml:"info-type,attr"`
	Data     string `xml:",innerxml"`
}

// NewCustomInfo factory for CustomInfo.
func NewCustomInfo(token xml.StartElement, reader xmlparse.TokenReader) (res CustomInfo, err error) {
	if res.Data, err = xmlparse.TokenRead(token.Name.Local, reader); err != nil {
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
