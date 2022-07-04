package entities

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/egnd/go-xmlparse"
	"github.com/egnd/go-xmlparse/fb2"
	"github.com/essentialkaos/translit/v2"
)

var (
	regexpYearPattern = regexp.MustCompile("[0-9]{3,}")
	currentYear       = time.Now().Year()
	booksNamePattern  = regexp.MustCompile(`[^a-zA-z0-9]+`)
	booksNamePartSize = 30
)

func ParseYear(date string) (res uint16) {
	if date == "" {
		return
	}

	for _, year := range regexpYearPattern.FindAllString(date, -1) {
		if len(year) <= 4 && !strings.HasPrefix(year, "0") {
			val, _ := strconv.ParseUint(year, 10, 16)
			res = uint16(val)

			if res > uint16(currentYear) {
				res = 0
			}

			break
		}
	}

	return
}

func SliceHasString(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}

func ParseFB2(reader io.Reader, encoder LibEncodeType,
	rules ...xmlparse.Rule,
) (res fb2.File, err error) {
	switch encoder {
	case LibEncodeParser:
		res, err = fb2.NewFile(xmlparse.NewDecoder(reader), rules...)
	default:
		err = xmlparse.NewDecoder(reader).Decode(&res)
	}

	return
}

func BuildBookURL(path, urlPrefix, pathPrefix string) string {
	return fmt.Sprintf("%s/%s",
		strings.Trim(urlPrefix, "/"),
		strings.Trim(strings.TrimPrefix(path, pathPrefix), "/"),
	)
}

func TransformStr(vals ...string) string {
	var buf bytes.Buffer

	for _, item := range vals {
		item = booksNamePattern.ReplaceAllString(strings.ToLower(translit.EncodeToICAO(item)), "-")
		if len(item) > booksNamePartSize {
			item = strings.Trim(item[0:booksNamePartSize], "-")
		}

		buf.WriteString(item)
	}

	return buf.String()
}

func BuildBookName(book *Book) (res string) {
	res = TransformStr(book.Info.Title)

	if authors := TransformStr(book.Info.Authors...); authors != "" {
		res += "." + authors
	}

	return
}
