package entities

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/egnd/go-fb2parse"
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

func GenerateID(args ...[]string) string {
	hasher := md5.New()

	for _, vals := range args {
		sort.Strings(vals)

		for _, str := range vals {
			str = strings.ToLower(strings.TrimSpace(str))

			if str != "" {
				hasher.Write([]byte(str))
			}
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func SliceHasString(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}

	return false
}
func GetFirstStr(items []string) string {
	for _, item := range items {
		if item != "" {
			return item
		}
	}

	return ""
}

func ParseFB2(reader io.Reader, encoder LibEncodeType,
	rules ...fb2parse.HandlingRule,
) (res fb2parse.FB2File, err error) {
	switch encoder {
	case LibEncodeParser:
		res, err = fb2parse.NewFB2File(fb2parse.NewDecoder(reader), rules...)
	default:
		err = fb2parse.NewDecoder(reader).Decode(&res)
	}

	return
}

func appendUniqStr(current *string, items ...string) {
	for _, item := range items {
		if item == "" || strings.Contains(*current, item) {
			continue
		}

		if *current == "" {
			*current = item
			continue
		}

		*current = fmt.Sprintf("%s%s%s", *current, indexFieldSep, item)
	}
}

func appendUniqFB2Author(current *string, items []fb2parse.FB2Author) {
	var buf bytes.Buffer
	var strVal string

	authors := make([]string, 0, len(items))

	for _, item := range items {
		if strVal = GetFirstStr(item.LastName); strVal != "" {
			buf.WriteString(strVal)
			buf.WriteRune(' ')
		}

		if strVal = GetFirstStr(item.FirstName); strVal != "" {
			buf.WriteString(strVal)
			buf.WriteRune(' ')
		}

		if strVal = GetFirstStr(item.MiddleName); strVal != "" {
			buf.WriteString(strVal)
			buf.WriteRune(' ')
		}

		if strVal = GetFirstStr(item.Nickname); strVal != "" {
			if buf.Len() > 0 {
				buf.WriteString("(")
				buf.WriteString(strVal)
				buf.WriteString(")")
			} else {
				buf.WriteString(strVal)
			}
		}

		authors = append(authors, string(bytes.TrimSpace(buf.Bytes())))
		buf.Reset()
	}

	appendUniqStr(current, authors...)
}

func appendUniqFB2Seq(current *string, items []fb2parse.FB2Sequence) {
	var buf bytes.Buffer

	seqs := make([]string, 0, len(items))

	for _, item := range items {
		if item.Name == "" {
			continue
		}

		buf.WriteString(item.Name)

		if item.Number != "" && item.Number != "0" {
			buf.WriteRune(' ')
			buf.WriteRune('(')
			buf.WriteString(item.Number)
			buf.WriteRune(')')

		}

		seqs = append(seqs, buf.String())
		buf.Reset()
	}

	appendUniqStr(current, seqs...)
}

func appendUniqFB2Publisher(current *string, items []fb2parse.FB2Publisher) {
	var buf bytes.Buffer
	var strVal string

	publ := make([]string, 0, len(items))

	for _, item := range items {
		if strVal = GetFirstStr(item.Publisher); strVal != "" {
			buf.WriteString(strVal)
		}

		if strVal = GetFirstStr(item.City); strVal != "" {
			buf.WriteRune(' ')

			if buf.Len() > 0 {
				buf.WriteRune('(')
				buf.WriteString(strVal)
				buf.WriteRune(')')
			} else {
				buf.WriteString(strVal)
			}
		}

		publ = append(publ, buf.String())
		buf.Reset()
	}

	appendUniqStr(current, publ...)
}

func BuildBookURL(path, urlPrefix, pathPrefix string) string {
	return fmt.Sprintf("%s/%s",
		strings.Trim(urlPrefix, "/"),
		strings.Trim(strings.TrimPrefix(path, pathPrefix), "/"),
	)
}

func TransformStr(val string) string {
	val = booksNamePattern.ReplaceAllString(
		strings.ToLower(
			translit.EncodeToICAO(strings.Split(val, indexFieldSep)[0]),
		),
		"-",
	)

	if len(val) > booksNamePartSize {
		val = strings.Trim(val[0:booksNamePartSize], "-")
	}

	return val
}

func BuildBookName(book BookIndex) (res string) {
	res = TransformStr(book.Titles)

	if authors := TransformStr(book.Authors); authors != "" {
		res += "." + authors
	}

	return
}
