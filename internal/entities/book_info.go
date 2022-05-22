package entities

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/egnd/fb2lib/pkg/fb2parser"
)

const (
	IndexFieldSep = "; "
)

type BookInfo struct {
	Offset           uint64    `json:"offset"`
	SizeCompressed   uint64    `json:"sizec"`
	SizeUncompressed uint64    `json:"size"`
	LibName          string    `json:"lib"`
	Src              string    `json:"src"`
	Index            BookIndex `json:"-"`
}

type BookIndex struct {
	Year      uint16
	ID        string
	Lang      string `json:"lng"`
	ISBN      string `json:"isbn"`
	Titles    string `json:"name"`
	Authors   string `json:"auth"`
	Sequences string `json:"seq"`
	Publisher string `json:"publ"`
	Date      string `json:"date"`
	Genres    string `json:"genr"`
}

func NewBookIndex(fb2 *fb2parser.FB2File) BookIndex {
	res := BookIndex{
		Titles: fb2.Description.TitleInfo.BookTitle,
		Date:   strings.TrimSpace(fb2.Description.TitleInfo.Date),
		Genres: strings.Join(fb2.Description.TitleInfo.Genre, IndexFieldSep),
		Lang:   fb2.Description.TitleInfo.Lang,
	}

	res.appendAuthors(fb2.Description.TitleInfo.Author)
	res.appendAuthors(fb2.Description.TitleInfo.Translator)
	res.appendSequences(fb2.Description.TitleInfo.Sequence)

	if fb2.Description.PublishInfo != nil {
		res.Publisher = fb2.Description.PublishInfo.Publisher
		if res.Publisher != "" && fb2.Description.PublishInfo.City != "" {
			res.Publisher += fmt.Sprintf(" (%s)", fb2.Description.PublishInfo.City)
		}

		res.ISBN = fb2.Description.PublishInfo.ISBN
		res.appendStr(fb2.Description.PublishInfo.BookName, &res.Titles)
		res.appendStr(fb2.Description.PublishInfo.Year, &res.Date)
	}

	if fb2.Description.SrcTitleInfo != nil {
		res.appendStr(fb2.Description.SrcTitleInfo.BookTitle, &res.Titles)
		res.appendAuthors(fb2.Description.SrcTitleInfo.Author)
		res.appendAuthors(fb2.Description.SrcTitleInfo.Translator)
		res.appendSequences(fb2.Description.SrcTitleInfo.Sequence)
		res.appendGenres(fb2.Description.SrcTitleInfo.Genre)
		res.appendStr(fb2.Description.SrcTitleInfo.Date, &res.Date)
	}

	res.Year = ParseYear(res.Date)
	res.ID = GenerateID([]string{res.ISBN, res.Lang, fmt.Sprint(res.Year)},
		strings.Split(res.Titles, ";"),
		strings.Split(strings.ReplaceAll(res.Authors, ",", ";"), ";"),
	)

	return res
}

func (bi *BookIndex) appendAuthors(items []fb2parser.FB2Author) {
	if len(items) == 0 {
		return
	}

	var buf bytes.Buffer
	buf.WriteString(bi.Authors)
	buf.WriteString(IndexFieldSep)

	for _, item := range items {
		if itemStr := strings.TrimSpace(fmt.Sprintf("%s %s %s",
			item.FirstName, item.MiddleName, item.LastName,
		)); itemStr != "" && !strings.Contains(bi.Authors, itemStr) {
			buf.WriteString(itemStr)
			buf.WriteString(", ")
		}
	}

	bi.Authors = strings.Trim(buf.String(), IndexFieldSep+",")
}

func (bi *BookIndex) appendSequences(items []fb2parser.FB2Sequence) {
	if len(items) == 0 {
		return
	}

	var buf bytes.Buffer
	buf.WriteString(bi.Sequences)
	buf.WriteString(IndexFieldSep)

	for _, item := range items {
		if item.Name != "" && !strings.Contains(bi.Sequences, item.Name) {
			buf.WriteString(item.Name)

			if item.Number != "" {
				buf.WriteString(" (")
				buf.WriteString(item.Number)
				buf.WriteString(")")
			}

			buf.WriteString(", ")
		}
	}

	bi.Sequences = strings.Trim(buf.String(), IndexFieldSep+", ")
}

func (bi *BookIndex) appendGenres(items []string) {
	if len(items) == 0 {
		return
	}

	var buf bytes.Buffer
	buf.WriteString(bi.Sequences)
	buf.WriteString(IndexFieldSep)

	for _, item := range items {
		if item != "" && !strings.Contains(bi.Sequences, item) {
			buf.WriteString(item)
			buf.WriteString(", ")
		}
	}

	bi.Genres = strings.Trim(buf.String(), IndexFieldSep+", ")
}

func (bi *BookIndex) appendStr(val string, orig *string) {
	if val == "" || strings.Contains(*orig, val) {
		return
	}

	if *orig == "" {
		*orig = val
	} else {
		*orig = fmt.Sprintf("%s%s%s", *orig, IndexFieldSep, val)
	}

}

func NewBookIndexMapping() *mapping.IndexMappingImpl {
	searchField := bleve.NewTextFieldMapping()

	sortField := bleve.NewNumericFieldMapping()
	sortField.IncludeInAll = false
	sortField.DocValues = false
	sortField.Index = true

	books := bleve.NewDocumentMapping()

	books.AddFieldMappingsAt("year", sortField)
	books.AddFieldMappingsAt("lng", searchField)
	books.AddFieldMappingsAt("isbn", searchField)
	books.AddFieldMappingsAt("name", searchField)
	books.AddFieldMappingsAt("auth", searchField)
	books.AddFieldMappingsAt("seq", searchField)
	books.AddFieldMappingsAt("publ", searchField)
	books.AddFieldMappingsAt("date", searchField)
	books.AddFieldMappingsAt("genr", searchField)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("books", books)
	mapping.DefaultType = "books"
	mapping.DefaultMapping = books

	return mapping
}
