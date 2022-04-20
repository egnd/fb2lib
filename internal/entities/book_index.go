package entities

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/egnd/fb2lib/pkg/fb2parser"
)

var (
	IndexFieldSep = "; "
)

type BookIndex struct {
	Year             uint16
	Offset           uint64
	SizeCompressed   uint64
	SizeUncompressed uint64
	ID               string
	Lang             string
	Src              string
	ISBN             string
	Titles           string
	Authors          string
	Sequences        string
	Publisher        string
	Date             string
	Genres           string
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

	hasher := md5.New()
	hasher.Write([]byte(res.ISBN))
	hasher.Write([]byte(res.Titles))
	hasher.Write([]byte(res.Authors))

	res.ID = hex.EncodeToString(hasher.Sum(nil))
	res.Year = parseYear(res.Date)

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
	strIndexedField := bleve.NewTextFieldMapping()

	strField := bleve.NewTextFieldMapping()
	strField.Index = false
	strField.IncludeInAll = false
	strField.IncludeTermVectors = false
	strField.DocValues = false

	numField := bleve.NewNumericFieldMapping()
	numField.Index = false
	numField.IncludeInAll = false
	numField.DocValues = false

	numSortField := *numField
	numSortField.Index = true
	numSortField.Store = false

	books := bleve.NewDocumentMapping()

	books.AddFieldMappingsAt("Year", &numSortField)
	books.AddFieldMappingsAt("Offset", numField)
	books.AddFieldMappingsAt("SizeCompressed", numField)
	books.AddFieldMappingsAt("SizeUncompressed", numField)
	books.AddFieldMappingsAt("Lang", strField)
	books.AddFieldMappingsAt("Src", strField)

	books.AddFieldMappingsAt("ISBN", strIndexedField)
	books.AddFieldMappingsAt("Titles", strIndexedField)
	books.AddFieldMappingsAt("Authors", strIndexedField)
	books.AddFieldMappingsAt("Sequences", strIndexedField)
	books.AddFieldMappingsAt("Publisher", strIndexedField)
	books.AddFieldMappingsAt("Date", strIndexedField)
	books.AddFieldMappingsAt("Genres", strIndexedField)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("books", books)
	mapping.DefaultType = "books"
	mapping.DefaultMapping = books

	return mapping
}
