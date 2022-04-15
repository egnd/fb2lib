package entities

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"gitlab.com/egnd/bookshelf/pkg/fb2parser"
)

var (
	IndexFieldSep = "; "
)

type BookIndex struct {
	ISBN      string
	Titles    string
	Authors   string
	Sequences string
	Publisher string
	Date      string
	Genres    string

	ID               string
	Src              string
	Offset           float64
	SizeCompressed   float64
	SizeUncompressed float64
}

func NewBookIndex(fb2 *fb2parser.FB2File) BookIndex {
	res := BookIndex{
		Titles: fb2.Description.TitleInfo.BookTitle,
		Date:   parseYear(fb2.Description.TitleInfo.Date),
		Genres: strings.Join(fb2.Description.TitleInfo.Genre, IndexFieldSep),
	}

	res.appendAuthors(fb2.Description.TitleInfo.Author)
	res.appendSequences(fb2.Description.TitleInfo.Sequence)

	if fb2.Description.PublishInfo != nil {
		res.ISBN = fb2.Description.PublishInfo.ISBN
		res.Publisher = fb2.Description.PublishInfo.Publisher
		res.appendStr(fb2.Description.PublishInfo.BookName, &res.Titles)
		res.appendStr(parseYear(fb2.Description.PublishInfo.Year), &res.Date)
	}

	if fb2.Description.SrcTitleInfo != nil {
		res.appendStr(fb2.Description.SrcTitleInfo.BookTitle, &res.Titles)
		res.appendAuthors(fb2.Description.SrcTitleInfo.Author)
		res.appendSequences(fb2.Description.SrcTitleInfo.Sequence)
		res.appendGenres(fb2.Description.SrcTitleInfo.Genre)
		res.appendStr(parseYear(fb2.Description.SrcTitleInfo.Date), &res.Date)
	}

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
			buf.WriteString(", ")
			buf.WriteString(itemStr)
		}
	}

	bi.Authors = strings.Trim(buf.String(), IndexFieldSep+", ")
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
			buf.WriteString(", ")
			buf.WriteString(item.Name)

			if item.Number != "" {
				buf.WriteString(" (")
				buf.WriteString(item.Number)
				buf.WriteString(")")
			}
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
			buf.WriteString(", ")
			buf.WriteString(item)
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
	strSearchField := bleve.NewTextFieldMapping()

	strField := bleve.NewTextFieldMapping()
	strField.Index = false
	strField.IncludeInAll = false
	strField.IncludeTermVectors = false
	strField.DocValues = false

	numField := bleve.NewNumericFieldMapping()
	numField.Index = false
	numField.IncludeInAll = false
	numField.DocValues = false

	books := bleve.NewDocumentMapping()
	books.AddFieldMappingsAt("ISBN", strSearchField)
	books.AddFieldMappingsAt("Titles", strSearchField)
	books.AddFieldMappingsAt("Authors", strSearchField)
	books.AddFieldMappingsAt("Sequences", strSearchField)
	books.AddFieldMappingsAt("Publisher", strSearchField)
	books.AddFieldMappingsAt("Date", strSearchField)
	books.AddFieldMappingsAt("Genres", strSearchField)
	books.AddFieldMappingsAt("Src", strField)
	books.AddFieldMappingsAt("Offset", numField)
	books.AddFieldMappingsAt("SizeCompressed", numField)
	books.AddFieldMappingsAt("SizeUncompressed", numField)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("books", books)
	mapping.DefaultType = "books"
	mapping.DefaultMapping = books

	return mapping
}
