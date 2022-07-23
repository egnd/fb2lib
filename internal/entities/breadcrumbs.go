package entities

type BreadCrumb struct {
	Title string
	Link  string
}

type BreadCrumbs []BreadCrumb

func (b BreadCrumbs) Push(title, link string) BreadCrumbs {
	return append(b, BreadCrumb{title, link})
}
