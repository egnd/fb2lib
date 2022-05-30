package fb2parse

import (
	"encoding/xml"
	"errors"
	"io"
)

// FB2Author struct of fb2 author.
// http://www.fictionbook.org/index.php/Элемент_author
type FB2Author struct {
	FirstName  []string `xml:"first-name"`
	MiddleName []string `xml:"middle-name"`
	LastName   []string `xml:"last-name"`
	Nickname   []string `xml:"nickname"`
	HomePage   []string `xml:"home-page"`
	Email      []string `xml:"email"`
	ID         []string `xml:"id"`
}

// NewFB2Author factory for FB2Author.
func NewFB2Author(tokenName string, reader xml.TokenReader) (res FB2Author, err error) { //nolint:gocognit,cyclop
	var token xml.Token

	var strVal string

loop:
	for {
		if token, err = reader.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}

			break
		}

		switch typedToken := token.(type) {
		case xml.StartElement:
			switch typedToken.Name.Local {
			case "first-name":
				if strVal, err = GetContent(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.FirstName = append(res.FirstName, strVal)
				}
			case "middle-name":
				if strVal, err = GetContent(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.MiddleName = append(res.MiddleName, strVal)
				}
			case "last-name":
				if strVal, err = GetContent(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.LastName = append(res.LastName, strVal)
				}
			case "nickname":
				if strVal, err = GetContent(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.Nickname = append(res.Nickname, strVal)
				}
			case "home-page":
				if strVal, err = GetContent(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.HomePage = append(res.HomePage, strVal)
				}
			case "email":
				if strVal, err = GetContent(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.Email = append(res.Email, strVal)
				}
			case "id":
				if strVal, err = GetContent(typedToken.Name.Local, reader); err == nil && strVal != "" {
					res.ID = append(res.ID, strVal)
				}
			}

			if err != nil {
				break loop
			}
		case xml.EndElement:
			if typedToken.Name.Local == tokenName {
				break loop
			}
		}
	}

	return res, err
}
