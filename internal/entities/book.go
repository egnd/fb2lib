package entities

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/egnd/go-xmlparse"
	"github.com/egnd/go-xmlparse/fb2"
)

type Book struct {
	ID             string          `json:"id"`
	Offset         uint64          `json:"from,omitempty"`
	Size           uint64          `json:"size,omitempty"`
	SizeCompressed uint64          `json:"sizec,omitempty"`
	Lib            string          `json:"lib,omitempty"`
	Src            string          `json:"src,omitempty"`
	Info           BookMeta        `json:"info,omitempty"`
	OrigInfo       *BookMeta       `json:"oinfo,omitempty"`
	PublInfo       []BookPublisher `json:"pinfo,omitempty"`
}

func (b *Book) Index() (res BookIndex) {
	var (
		isbn       bytes.Buffer
		title      bytes.Buffer
		author     bytes.Buffer
		translator bytes.Buffer
		serie      bytes.Buffer
		date       bytes.Buffer
		genre      bytes.Buffer
		publisher  bytes.Buffer
	)

	appendStr := func(buf *bytes.Buffer, data ...string) {
		for _, str := range data {
			if str == "" || bytes.Contains(buf.Bytes(), []byte(str)) {
				continue
			}
			buf.WriteString(str)
			buf.WriteRune(' ')
		}
	}

	appendStr(&title, b.Info.Title)
	appendStr(&date, b.Info.Date)
	appendStr(&genre, b.Info.Genres...)
	appendStr(&author, b.Info.Authors...)
	appendStr(&translator, b.Info.Translators...)
	appendStr(&serie, b.Info.Sequences...)

	if b.OrigInfo != nil {
		appendStr(&title, b.OrigInfo.Title)
		appendStr(&date, b.OrigInfo.Date)
		appendStr(&genre, b.OrigInfo.Genres...)
		appendStr(&author, b.OrigInfo.Authors...)
		appendStr(&translator, b.OrigInfo.Translators...)
		appendStr(&serie, b.OrigInfo.Sequences...)
	}

	for _, publ := range b.PublInfo {
		appendStr(&title, publ.Title)
		appendStr(&date, publ.Year)
		appendStr(&author, publ.Authors...)
		appendStr(&serie, publ.Sequences...)
		appendStr(&publisher, publ.Publisher)
		appendStr(&isbn, publ.ISBN)
	}

	res.ID = b.ID
	res.Lang = b.Info.Lang
	res.Lib = b.Lib
	res.ISBN = isbn.String()
	res.Title = title.String()
	res.Author = author.String()
	res.Translator = translator.String()
	res.Serie = serie.String()
	res.Date = date.String()
	res.Genre = genre.String()
	res.Publisher = publisher.String()
	res.Year = ParseYear(res.Date)

	return
}

func (b *Book) Authors() []string {
	index := make(map[string]struct{}, 6)

	for _, item := range b.Info.Authors {
		index[item] = struct{}{}
	}

	if b.OrigInfo != nil {
		for _, item := range b.OrigInfo.Authors {
			index[item] = struct{}{}
		}
	}

	for _, item := range b.PublInfo {
		for _, item := range item.Authors {
			index[item] = struct{}{}
		}
	}

	res := make([]string, 0, len(index))
	for k := range index {
		res = append(res, k)
	}

	return res
}

func (b *Book) Translators() []string {
	index := make(map[string]struct{}, 6)

	for _, item := range b.Info.Translators {
		index[item] = struct{}{}
	}

	if b.OrigInfo != nil {
		for _, item := range b.OrigInfo.Translators {
			index[item] = struct{}{}
		}
	}

	res := make([]string, 0, len(index))
	for k := range index {
		res = append(res, k)
	}

	return res
}

func (b *Book) Genres() []string {
	index := make(map[string]struct{}, 6)

	for _, item := range b.Info.Genres {
		index[item] = struct{}{}
	}

	if b.OrigInfo != nil {
		for _, item := range b.OrigInfo.Genres {
			index[item] = struct{}{}
		}
	}

	res := make([]string, 0, len(index))
	for k := range index {
		res = append(res, k)
	}

	return res
}

func (b *Book) Series() []string {
	index := make(map[string]struct{}, 6)

	for _, item := range b.Info.Sequences {
		index[item] = struct{}{}
	}

	if b.OrigInfo != nil {
		for _, item := range b.OrigInfo.Sequences {
			index[item] = struct{}{}
		}
	}

	for _, item := range b.PublInfo {
		for _, item := range item.Sequences {
			index[item] = struct{}{}
		}
	}

	res := make([]string, 0, len(index))
	for k := range index {
		res = append(res, k)
	}

	return res
}

func (b *Book) Titles() []string {
	index := make(map[string]struct{}, 6)

	index[b.Info.Title] = struct{}{}

	if b.OrigInfo != nil {
		index[b.OrigInfo.Title] = struct{}{}
	}

	for _, item := range b.PublInfo {
		index[item.Title] = struct{}{}
	}

	res := make([]string, 0, len(index))
	for k := range index {
		res = append(res, k)
	}

	return res
}

func (b *Book) ReadFB2(data *fb2.File) {
	var misc []string

	for k, item := range data.Description {
		if k == 0 {
			b.Info = NewBookMeta(item.TitleInfo, data.Binary)
			misc = append(misc, b.Info.Lang)
		}

		if b.OrigInfo == nil && len(item.SrcTitleInfo) > 0 {
			orig := NewBookMeta(item.SrcTitleInfo, data.Binary)
			b.OrigInfo = &orig
		}

		for _, publ := range item.PublishInfo {
			publInfo := NewBookPublisher(publ)
			b.PublInfo = append(b.PublInfo, publInfo)
			misc = append(misc, publInfo.ISBN)
		}
	}

	hasher := md5.New()
	for _, vals := range [][]string{misc, b.Titles(), b.Authors(), b.Translators()} {
		sort.Strings(vals)
		for _, str := range vals {
			str = strings.ToLower(strings.TrimSpace(str))
			if str != "" {
				hasher.Write([]byte(str))
			}
		}
	}

	b.ID = hex.EncodeToString(hasher.Sum(nil))
}

type BookMeta struct {
	Annotation  string      `json:"annot,omitempty"`
	Lang        string      `json:"lang,omitempty"`
	SrcLang     string      `json:"slang,omitempty"`
	Title       string      `json:"title,omitempty"`
	Keywords    string      `json:"kwds,omitempty"`
	Date        string      `json:"date,omitempty"`
	Genres      []string    `json:"genres,omitempty"`
	Authors     []string    `json:"auth,omitempty"`
	Translators []string    `json:"transl,omitempty"`
	Sequences   []string    `json:"seq,omitempty"`
	CoverID     string      `json:"cover,omitempty"`
	Cover       *fb2.Binary `json:"-"`
}

func NewBookMeta(data []fb2.TitleInfo, bin []fb2.Binary) (res BookMeta) {
	for _, item := range data {
		res.Genres = append(res.Genres, item.Genre...)

		if res.Annotation == "" {
			for _, annot := range item.Annotation {
				res.Annotation = annot.HTML
				break
			}
		}

		for _, v := range item.Author {
			res.Authors = append(res.Authors, v.String())
		}

		for _, v := range item.Translator {
			res.Translators = append(res.Translators, v.String())
		}

		for _, v := range item.Sequence {
			res.Sequences = append(res.Sequences, v.String())
		}

		if res.Title == "" {
			res.Title = xmlparse.GetStrFrom(item.BookTitle)
		}
		if res.Keywords == "" {
			res.Keywords = xmlparse.GetStrFrom(item.Keywords)
		}
		if res.Date == "" {
			res.Date = xmlparse.GetStrFrom(item.Date)
		}

		if res.Lang == "" {
			res.Lang = xmlparse.GetStrFrom(item.Lang)
		}

		if res.SrcLang == "" {
			res.SrcLang = xmlparse.GetStrFrom(item.SrcLang)
		}

		if res.CoverID != "" {
			continue
		}

	loop:
		for _, cover := range item.Coverpage {
			for _, img := range cover.Images {
				res.CoverID = strings.TrimPrefix(img.Href, "#")
				break loop
			}
		}
	}

	return
}

type BookPublisher struct {
	Title     string   `json:"title,omitempty"`
	Publisher string   `json:"publ,omitempty"`
	Year      string   `json:"year,omitempty"`
	ISBN      string   `json:"isbn,omitempty"`
	Authors   []string `json:"auth,omitempty"`
	Sequences []string `json:"seqs,omitempty"`
}

func NewBookPublisher(data fb2.Publisher) (res BookPublisher) {
	res.Title = xmlparse.GetStrFrom(data.BookName)
	res.Publisher = data.String()
	res.Year = xmlparse.GetStrFrom(data.Year)
	res.ISBN = xmlparse.GetStrFrom(data.ISBN)
	res.Authors = append(res.Authors, data.BookAuthor...)

	for _, v := range data.Sequence {
		res.Sequences = append(res.Sequences, v.String())
	}

	return
}

type IndexField string

const (
	IdxFUndefined  IndexField = ""
	IdxFID         IndexField = "id"
	IdxFYear       IndexField = "year"
	IdxFISBN       IndexField = "isbn"
	IdxFTitle      IndexField = "title"
	IdxFAuthor     IndexField = "auth"
	IdxFTranslator IndexField = "transl"
	IdxFSerie      IndexField = "seq"
	IdxFDate       IndexField = "date"
	IdxFGenre      IndexField = "genre"
	IdxFPublisher  IndexField = "publ"
	IdxFLang       IndexField = "lng"
	IdxFKeywords   IndexField = "kwds"
	IdxFLib        IndexField = "lib"
)

type BookIndex struct {
	ID         string `json:"id,omitempty"`
	Year       uint16 `json:"year,omitempty"`
	ISBN       string `json:"isbn,omitempty"`
	Title      string `json:"title,omitempty"`
	Author     string `json:"auth,omitempty"`
	Translator string `json:"transl,omitempty"`
	Serie      string `json:"seq,omitempty"`
	Date       string `json:"date,omitempty"`
	Genre      string `json:"genre,omitempty"`
	Publisher  string `json:"publ,omitempty"`
	Lang       string `json:"lng,omitempty"`
	Keywords   string `json:"kwds,omitempty"`
	Lib        string `json:"lib,omitempty"`
}

func NewBookIndexMapping() *mapping.IndexMappingImpl {
	books := bleve.NewDocumentMapping()

	sortField := bleve.NewNumericFieldMapping()
	sortField.IncludeInAll = false
	sortField.DocValues = false
	books.AddFieldMappingsAt(string(IdxFID), sortField)

	strField := bleve.NewTextFieldMapping()
	strField.Store = false
	books.AddFieldMappingsAt(string(IdxFISBN), strField)
	books.AddFieldMappingsAt(string(IdxFTitle), strField)
	books.AddFieldMappingsAt(string(IdxFAuthor), strField)
	books.AddFieldMappingsAt(string(IdxFTranslator), strField)
	books.AddFieldMappingsAt(string(IdxFSerie), strField)
	books.AddFieldMappingsAt(string(IdxFDate), strField)
	books.AddFieldMappingsAt(string(IdxFGenre), strField)
	books.AddFieldMappingsAt(string(IdxFPublisher), strField)
	books.AddFieldMappingsAt(string(IdxFLang), strField)
	books.AddFieldMappingsAt(string(IdxFKeywords), strField)
	books.AddFieldMappingsAt(string(IdxFLib), strField)

	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("books", books)
	mapping.DefaultType = "books"
	mapping.DefaultMapping = books

	return mapping
}
