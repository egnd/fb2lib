package entities

import (
	"bytes"
	"fmt"
	"strings"

	"gitlab.com/egnd/bookshelf/pkg/fb2parser"
)

type BookIndex struct {
	ID        string
	ISBN      string
	Titles    string
	Authors   string
	Sequences string
	Date      string
	Publisher string
	Src       string
	Offset    uint64
	Size      uint64
}

func (bi *BookIndex) appendAuthors(items []fb2parser.FB2Author) {
	if len(items) == 0 {
		return
	}

	var buf bytes.Buffer
	buf.WriteString(bi.Authors)

	for _, item := range items {
		if itemStr := strings.TrimSpace(fmt.Sprintf("%s %s %s",
			item.FirstName, item.MiddleName, item.LastName,
		)); itemStr != "" && !strings.Contains(bi.Authors, itemStr) {
			buf.WriteString(", ")
			buf.WriteString(itemStr)
		}
	}

	bi.Authors = strings.TrimPrefix(buf.String(), ", ")
}

func (bi *BookIndex) appendSequences(items []fb2parser.FB2Sequence) {
	if len(items) == 0 {
		return
	}

	var buf bytes.Buffer
	buf.WriteString(bi.Sequences)

	for _, item := range items {
		if item.Name != "" && !strings.Contains(bi.Sequences, item.Name) {
			buf.WriteString(", ")
			buf.WriteString(item.Name)

			if item.Number > 0 {
				buf.WriteString(" (")
				buf.WriteString(fmt.Sprint(item.Number))
				buf.WriteString(")")
			}
		}
	}

	bi.Sequences = strings.TrimPrefix(buf.String(), ", ")
}

func (bi *BookIndex) appendStr(val string, orig *string) {
	if val == "" || strings.Contains(*orig, val) {
		return
	}

	if *orig == "" {
		*orig = val
	} else {
		*orig += ", " + val
	}

}

func NewBookIndex(fb2 *fb2parser.FB2File) BookIndex {
	res := BookIndex{
		Titles: fb2.Description.TitleInfo.BookTitle,
	}

	res.appendAuthors(fb2.Description.TitleInfo.Author)
	res.appendSequences(fb2.Description.TitleInfo.Sequence)

	if fb2.Description.PublishInfo != nil {
		res.ISBN = fb2.Description.PublishInfo.ISBN
		res.Publisher = fb2.Description.PublishInfo.Publisher
		res.appendStr(fb2.Description.PublishInfo.BookName, &res.Titles)

		if fb2.Description.PublishInfo.Year > 0 {
			res.Date = fmt.Sprint(fb2.Description.PublishInfo.Year)
		}
	}

	if fb2.Description.SrcTitleInfo != nil {
		res.appendStr(fb2.Description.SrcTitleInfo.BookTitle, &res.Titles)
		res.appendAuthors(fb2.Description.SrcTitleInfo.Author)
		res.appendSequences(fb2.Description.SrcTitleInfo.Sequence)

		if res.Date == "" {
			res.Date = fb2.Description.SrcTitleInfo.Date
		}
	}

	return res
}
