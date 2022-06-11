package entities

import (
	"strings"

	"github.com/egnd/go-xmlparse/fb2"
)

type BookInfo struct {
	Offset         uint64    `json:"from"`
	Size           uint64    `json:"size"`
	SizeCompressed uint64    `json:"sizec"`
	LibName        string    `json:"lib"`
	Src            string    `json:"src"`
	Index          BookIndex `json:"-"`
	Details        struct {
		Images     []fb2.Binary
		Annotation string
	} `json:"-"`
}

func (b *BookInfo) ReadDetails(fb2File *fb2.File) {
	b.Details.Images = make([]fb2.Binary, 0, len(fb2File.Binary))
	index := make(map[string]*fb2.Binary, len(fb2File.Binary))

	for k := range fb2File.Binary {
		index[fb2File.Binary[k].ID] = &fb2File.Binary[k]
	}

	for _, descr := range fb2File.Description {
		for _, title := range descr.TitleInfo {
			for _, cover := range title.Coverpage {
				for _, img := range cover.Images {
					if _, ok := index[strings.TrimPrefix(img.Href, "#")]; ok {
						b.Details.Images = append(b.Details.Images, *index[strings.TrimPrefix(img.Href, "#")])
					}
				}
			}

			for _, annot := range title.Annotation {
				b.Details.Annotation += annot.HTML
			}
		}

		for _, title := range descr.SrcTitleInfo {
			for _, cover := range title.Coverpage {
				for _, img := range cover.Images {
					if _, ok := index[strings.TrimPrefix(img.Href, "#")]; ok {
						b.Details.Images = append(b.Details.Images, *index[strings.TrimPrefix(img.Href, "#")])
					}
				}
			}

			if b.Details.Annotation != "" {
				for _, annot := range title.Annotation {
					b.Details.Annotation += annot.HTML
				}
			}
		}
	}
}
