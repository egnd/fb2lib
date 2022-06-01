package entities

import (
	"strings"

	"github.com/egnd/go-fb2parse"
)

type BookInfo struct {
	Offset         uint64    `json:"from"`
	Size           uint64    `json:"size"`
	SizeCompressed uint64    `json:"sizec"`
	LibName        string    `json:"lib"`
	Src            string    `json:"src"`
	Index          BookIndex `json:"-"`
	Details        struct {
		Images     []fb2parse.FB2Binary
		Annotation string
	} `json:"-"`
}

func (b *BookInfo) ReadDetails(fb2 *fb2parse.FB2File) {
	b.Details.Images = make([]fb2parse.FB2Binary, 0, len(fb2.Binary))
	index := make(map[string]*fb2parse.FB2Binary, len(fb2.Binary))

	for k := range fb2.Binary {
		index[fb2.Binary[k].ID] = &fb2.Binary[k]
	}

	for _, descr := range fb2.Description {
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
