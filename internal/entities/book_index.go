package entities

import (
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/egnd/go-xmlparse/fb2"
)

const (
	IndexFieldSep = "; "

	IdxFieldYear       = "year"
	IdxFieldISBN       = "isbn"
	IdxFieldTitle      = "title"
	IdxFieldAuthor     = "author"
	IdxFieldTranslator = "transl"
	IdxFieldSerie      = "serie"
	IdxFieldDate       = "date"
	IdxFieldGenre      = "genre"
	IdxFieldPublisher  = "publ"
	IdxFieldLang       = "lang"
	IdxFieldLib        = "lib"
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

func NewFB2Index(fb2 *fb2.File) (res BookIndex) {
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
		ISBN:       getVal(IdxFieldISBN),
		Title:      getVal(IdxFieldTitle),
		Author:     getVal(IdxFieldAuthor),
		Translator: getVal(IdxFieldTranslator),
		Serie:      getVal(IdxFieldSerie),
		Date:       getVal(IdxFieldDate),
		Genre:      getVal(IdxFieldGenre),
		Publisher:  getVal(IdxFieldPublisher),
		Lang:       getVal(IdxFieldLang),
		Lib:        getVal(IdxFieldLib),
	}
}

func NewBookIndexMapping() *mapping.IndexMappingImpl {
	books := bleve.NewDocumentMapping()

	sortField := bleve.NewNumericFieldMapping()
	sortField.IncludeInAll = false
	sortField.DocValues = false
	sortField.Index = true
	books.AddFieldMappingsAt(IdxFieldYear, sortField)

	searchField := bleve.NewTextFieldMapping()
	books.AddFieldMappingsAt(IdxFieldISBN, searchField)
	books.AddFieldMappingsAt(IdxFieldTitle, searchField)
	books.AddFieldMappingsAt(IdxFieldAuthor, searchField)
	books.AddFieldMappingsAt(IdxFieldTranslator, searchField)
	books.AddFieldMappingsAt(IdxFieldSerie, searchField)
	books.AddFieldMappingsAt(IdxFieldDate, searchField)
	books.AddFieldMappingsAt(IdxFieldGenre, searchField)
	books.AddFieldMappingsAt(IdxFieldPublisher, searchField)
	books.AddFieldMappingsAt(IdxFieldLang, searchField)
	books.AddFieldMappingsAt(IdxFieldLib, searchField)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("books", books)
	mapping.DefaultType = "books"
	mapping.DefaultMapping = books

	return mapping
}
