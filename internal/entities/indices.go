package entities

import (
	"fmt"
	"strings"
)

type BookIndex struct {
	ID        string
	ISBN      string
	Titles    string
	Authors   string
	Sequences string
	Genres    string
	Date      string
	Lang      string
	// Annotation string
	Archive        string
	Offset         int64
	SizeCompressed int64
}

func NewBookIndexFrom(file *FB2File) BookIndex {
	res := BookIndex{
		Titles: file.Description.TitleInfo.BookTitle,
		Authors: func() string {
			items := make([]string, 0, len(file.Description.TitleInfo.Author))
			for _, item := range file.Description.TitleInfo.Author {
				items = append(items, strings.TrimSpace(fmt.Sprintf("%s %s %s", item.FirstName, item.MiddleName, item.LastName)))
			}
			return strings.Join(items, ", ")
		}(),
		Sequences: func() string {
			items := make([]string, 0, len(file.Description.TitleInfo.Sequence))
			for _, item := range file.Description.TitleInfo.Sequence {
				items = append(items, item.Name)
			}
			return strings.Join(items, ", ")
		}(),
		Genres: strings.Join(file.Description.TitleInfo.Genre, ", "),
		Date:   file.Description.TitleInfo.Date,
		Lang:   file.Description.TitleInfo.Lang,
		// Annotation: file.Description.TitleInfo.Annotation.HTML,
	}

	var author string
	if file.Description.SrcTitleInfo != nil {
		res.Titles += ", " + file.Description.SrcTitleInfo.BookTitle

		items := make([]string, 0, len(file.Description.SrcTitleInfo.Author))
		for _, item := range file.Description.SrcTitleInfo.Author {
			if author = strings.TrimSpace(fmt.Sprintf("%s %s %s", item.FirstName, item.MiddleName, item.LastName)); author == "" {
				continue
			}
			items = append(items, author)
		}
		if len(items) > 0 {
			res.Authors += ", " + strings.Join(items, ", ")
		}

		items = make([]string, 0, len(file.Description.SrcTitleInfo.Sequence))
		for _, item := range file.Description.SrcTitleInfo.Sequence {
			if item.Name == "" {
				continue
			}
			items = append(items, item.Name)
		}
		if len(items) > 0 {
			res.Sequences += ", " + strings.Join(items, ", ")
		}

		if res.Genres == "" {
			res.Genres = strings.Join(file.Description.SrcTitleInfo.Genre, ", ")
		}

		if res.Date == "" {
			res.Date = file.Description.SrcTitleInfo.Date
		}
		// if res.Annotation == "" {
		// 	res.Annotation = file.Description.SrcTitleInfo.Annotation.HTML
		// }
	}

	if file.Description.PublishInfo != nil {
		if file.Description.PublishInfo.BookName != "" {
			res.Titles += ", " + file.Description.PublishInfo.BookName
		}
		if file.Description.PublishInfo.ISBN != "" {
			res.ISBN += ", " + file.Description.PublishInfo.ISBN
		}
	}

	return res
}
