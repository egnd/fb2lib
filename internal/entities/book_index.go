package entities

import (
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/egnd/go-fb2parse"
)

const (
	IndexFieldSep = "; "
)

type BookIndex struct {
	Year       uint16 `json:"year"`
	ID         string
	ISBN       string `json:"isbn"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	Translator string `json:"transl"`
	Serie      string `json:"serie"`
	Date       string `json:"date"`
	Genre      string `json:"genre"`
	Publisher  string `json:"publ"`
	Lang       string `json:"lang"`
	Lib        string `json:"lib"`
}

func NewFB2Index(fb2 *fb2parse.FB2File) (res BookIndex) {
	for _, descr := range fb2.Description {
		for _, title := range descr.TitleInfo {
			appendUniqStr(&res.Title, title.BookTitle...)
			appendUniqFB2Author(&res.Author, title.Author)
			appendUniqFB2Author(&res.Translator, title.Translator)
			appendUniqFB2Seq(&res.Serie, title.Sequence)
			appendUniqStr(&res.Date, title.Date...)
			appendUniqStr(&res.Genre, title.Genre...)
			appendUniqStr(&res.Lang, title.Lang...)
		}

		for _, srcTitle := range descr.SrcTitleInfo {
			appendUniqStr(&res.Title, srcTitle.BookTitle...)
			appendUniqFB2Author(&res.Author, srcTitle.Author)
			appendUniqFB2Author(&res.Translator, srcTitle.Translator)
			appendUniqFB2Seq(&res.Serie, srcTitle.Sequence)
			appendUniqStr(&res.Date, srcTitle.Date...)
			appendUniqStr(&res.Genre, srcTitle.Genre...)
		}

		for _, publish := range descr.PublishInfo {
			appendUniqStr(&res.ISBN, publish.ISBN...)
			appendUniqStr(&res.Title, publish.BookName...)
			appendUniqStr(&res.Date, publish.Year...)
		}

		appendUniqFB2Publisher(&res.Publisher, descr.PublishInfo)
	}

	res.Year = ParseYear(res.Date)
	res.ID = GenerateID(
		[]string{res.ISBN, res.Lang},
		strings.Split(res.Title, IndexFieldSep),
		strings.Split(res.Author, IndexFieldSep),
		strings.Split(res.Translator, IndexFieldSep),
	)

	return
}

func NewBookIndex(match *search.DocumentMatch) BookIndex {
	getVal := func(fieldName string) string {
		if highlights, ok := match.Fragments[fieldName]; ok && len(highlights) > 0 {
			return highlights[0]
		}

		if val, ok := match.Fields[fieldName]; ok {
			return val.(string)
		}

		return ""
	}

	return BookIndex{
		ID:         match.ID,
		ISBN:       getVal("isbn"),
		Title:      getVal("title"),
		Author:     getVal("author"),
		Translator: getVal("transl"),
		Serie:      getVal("serie"),
		Date:       getVal("date"),
		Genre:      getVal("genre"),
		Publisher:  getVal("publ"),
		Lang:       getVal("lang"),
		Lib:        getVal("lib"),
	}
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
	books.AddFieldMappingsAt("title", searchField)
	books.AddFieldMappingsAt("author", searchField)
	books.AddFieldMappingsAt("transl", searchField)
	books.AddFieldMappingsAt("serie", searchField)
	books.AddFieldMappingsAt("date", searchField)
	books.AddFieldMappingsAt("genre", searchField)
	books.AddFieldMappingsAt("publ", searchField)
	books.AddFieldMappingsAt("lang", searchField)
	books.AddFieldMappingsAt("lib", searchField)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("books", books)
	mapping.DefaultType = "books"
	mapping.DefaultMapping = books

	return mapping
}
