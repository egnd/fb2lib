package response

import (
	"fmt"
	"regexp"
	"strings"

	translit "github.com/essentialkaos/translit/v2"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

var (
	booksNamePattern  = regexp.MustCompile(`[^a-zA-z0-9]+`)
	booksNamePartSize = 30
)

func BuildBookURL(path, urlPrefix, pathPrefix string) string {
	return fmt.Sprintf("%s/%s",
		strings.Trim(urlPrefix, "/"),
		strings.Trim(strings.TrimPrefix(path, pathPrefix), "/"),
	)
}

func TransformStr(val string) string {
	val = booksNamePattern.ReplaceAllString(
		strings.ToLower(
			translit.EncodeToICAO(strings.Split(val, ",")[0]),
		),
		"-",
	)

	if len(val) > booksNamePartSize {
		val = strings.Trim(val[0:booksNamePartSize], "-")
	}

	return val
}

func BuildBookName(book entities.BookIndex) (res string) {
	res = TransformStr(book.Titles)

	if authors := TransformStr(book.Authors); authors != "" {
		res += "." + authors
	}

	return
}
