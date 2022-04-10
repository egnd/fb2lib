package fb2parser

import (
	"encoding/xml"
	"strings"
)

func GetTokenValue(curTag string, reader xml.TokenReader) string {
	var buf strings.Builder

	for {
		token, err := reader.Token()
		if err != nil {
			return strings.TrimSpace(buf.String())
		}

		switch typedToken := token.(type) {
		case xml.CharData:
			buf.Write(typedToken)
		case xml.EndElement:
			if typedToken.Name.Local == curTag {
				return strings.TrimSpace(buf.String())
			}
		}
	}
}
