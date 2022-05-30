package entities

import (
	"fmt"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/egnd/go-fb2parse"
)

const (
	indexFieldSep = "; "
)

type BookIndex struct {
	Year      uint16 `json:"year"`
	ID        string
	ISBN      string `json:"isbn"`
	Titles    string `json:"name"`
	Authors   string `json:"auth"`
	Sequences string `json:"seq"`
	Date      string `json:"date"`
	Genres    string `json:"genr"`
	Publisher string `json:"publ"`
	Lang      string `json:"lng"`
}

func NewFB2Index(fb2 *fb2parse.FB2File) (res BookIndex) {
	for _, descr := range fb2.Description {
		for _, title := range descr.TitleInfo {
			appendUniqStr(&res.Titles, title.BookTitle...)
			appendUniqFB2Author(&res.Authors, title.Author)
			appendUniqFB2Author(&res.Authors, title.Translator)
			appendUniqFB2Seq(&res.Sequences, title.Sequence)
			appendUniqStr(&res.Date, title.Date...)
			appendUniqStr(&res.Genres, title.Genre...)
			appendUniqStr(&res.Lang, title.Lang...)
		}

		for _, srcTitle := range descr.SrcTitleInfo {
			appendUniqStr(&res.Titles, srcTitle.BookTitle...)
			appendUniqFB2Author(&res.Authors, srcTitle.Author)
			appendUniqFB2Seq(&res.Sequences, srcTitle.Sequence)
			appendUniqStr(&res.Date, srcTitle.Date...)
			appendUniqStr(&res.Genres, srcTitle.Genre...)
		}

		for _, publish := range descr.PublishInfo {
			appendUniqStr(&res.ISBN, publish.ISBN...)
			appendUniqStr(&res.Titles, publish.BookName...)
			appendUniqStr(&res.Date, publish.Year...)

		}

		appendUniqFB2Publisher(&res.Publisher, descr.PublishInfo)
	}

	res.Year = ParseYear(res.Date)
	res.ID = GenerateID(
		[]string{res.ISBN, res.Lang, fmt.Sprint(res.Year)},
		strings.Split(res.Titles, indexFieldSep),
		strings.Split(res.Authors, indexFieldSep),
	)

	return
}

func NewBookIndexMapping() *mapping.IndexMappingImpl {
	books := bleve.NewDocumentMapping()

	sortField := bleve.NewNumericFieldMapping()
	sortField.IncludeInAll = false
	sortField.DocValues = false
	sortField.Index = true
	books.AddFieldMappingsAt("year", sortField)

	searchField := bleve.NewTextFieldMapping()
	books.AddFieldMappingsAt("isbn", searchField)
	books.AddFieldMappingsAt("name", searchField)
	books.AddFieldMappingsAt("auth", searchField)
	books.AddFieldMappingsAt("seq", searchField)
	books.AddFieldMappingsAt("date", searchField)
	books.AddFieldMappingsAt("genr", searchField)
	books.AddFieldMappingsAt("publ", searchField)
	books.AddFieldMappingsAt("lng", searchField)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("books", books)
	mapping.DefaultType = "books"
	mapping.DefaultMapping = books

	return mapping
}
