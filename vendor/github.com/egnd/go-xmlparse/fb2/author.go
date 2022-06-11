package fb2

import (
	"encoding/xml"

	"github.com/egnd/go-xmlparse"
)

// Author struct of fb2 author.
// http://www.fictionbook.org/index.php/Элемент_author
type Author struct {
	FirstName  []string `xml:"first-name"`
	MiddleName []string `xml:"middle-name"`
	LastName   []string `xml:"last-name"`
	Nickname   []string `xml:"nickname"`
	HomePage   []string `xml:"home-page"`
	Email      []string `xml:"email"`
	ID         []string `xml:"id"`
}

// NewAuthor factory for Author.
func NewAuthor(tokenName string, reader xmlparse.TokenReader) (res Author, err error) { //nolint:gocognit,cyclop
	var token xml.Token

	var strVal string

	for {
		if token, err = reader.Token(); err != nil {
			return
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			switch typedToken.Name.Local {
			case "first-name":
				if strVal, err = xmlparse.TokenRead(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.FirstName = append(res.FirstName, strVal)
				}
			case "middle-name":
				if strVal, err = xmlparse.TokenRead(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.MiddleName = append(res.MiddleName, strVal)
				}
			case "last-name":
				if strVal, err = xmlparse.TokenRead(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.LastName = append(res.LastName, strVal)
				}
			case "nickname":
				if strVal, err = xmlparse.TokenRead(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.Nickname = append(res.Nickname, strVal)
				}
			case "home-page":
				if strVal, err = xmlparse.TokenRead(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.HomePage = append(res.HomePage, strVal)
				}
			case "email":
				if strVal, err = xmlparse.TokenRead(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.Email = append(res.Email, strVal)
				}
			case "id":
				if strVal, err = xmlparse.TokenRead(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.ID = append(res.ID, strVal)
				}
			}

			if err != nil {
				return
			}
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				return res, nil
			}
		}
	}
}
