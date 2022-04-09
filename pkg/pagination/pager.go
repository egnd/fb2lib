package pagination

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type Pager struct {
	curPage     int
	pageSize    int
	pagesCnt    int
	itemsCnt    uint64
	curPageKey  string
	pageSizeKey string
	req         *http.Request
}

func NewPager(req *http.Request) *Pager {
	return &Pager{
		curPageKey:  "page",
		pageSizeKey: "per",
		req:         req,
	}
}

func (p *Pager) getFromReqInt(key string, def int) int {
	val := p.req.URL.Query().Get(key)
	if val == "" {
		return def
	}

	res, err := strconv.Atoi(val)
	if err != nil || res == 0 {
		return def
	}

	return res
}

func (p *Pager) SetCurPage(val int, key ...string) IPager {
	if len(key) > 0 && key[0] != "" {
		p.curPageKey = key[0]
	}

	p.curPage = val

	return p
}

func (p *Pager) ReadCurPage(key ...string) IPager {
	if len(key) > 0 && key[0] != "" {
		p.curPageKey = key[0]
	}

	p.curPage = p.getFromReqInt(p.curPageKey, 1)

	return p
}

func (p *Pager) GetCurPage() int {
	if p.curPage < 1 {
		return 1
	}

	return p.curPage
}

func (p *Pager) IsCurPage(val int) bool {
	return p.GetCurPage() == val
}

func (p *Pager) ReadPageSize(key ...string) IPager {
	if len(key) > 0 && key[0] != "" {
		p.pageSizeKey = key[0]
	}

	p.pageSize = p.getFromReqInt(p.pageSizeKey, p.pageSize)

	return p
}

func (p *Pager) SetPageSize(val int, key ...string) IPager {
	if len(key) > 0 && key[0] != "" {
		p.pageSizeKey = key[0]
	}

	p.pageSize = val

	return p
}

func (p *Pager) GetPageSize() int {
	if p.pageSize < 1 {
		return 10
	}

	return p.pageSize
}

func (p *Pager) GetOffset() int {
	return (p.GetCurPage() - 1) * p.pageSize
}

func (p *Pager) SetTotal(val interface{}) IPager {
	p.itemsCnt, _ = toUInt64(val)

	return p
}

func (p *Pager) GetTotal() uint64 {
	return p.itemsCnt
}

func (p *Pager) GetPagesCnt() int {
	return int(math.Ceil(float64(p.itemsCnt) / float64(p.pageSize)))
}

func (p *Pager) HasPages() bool {
	return p.GetPagesCnt() > 1
}

func (p *Pager) HasPrev() bool {
	return p.GetCurPage() > 1
}

func (p *Pager) HasNext() bool {
	return p.GetCurPage() < p.GetPagesCnt()
}

func (p *Pager) GetPages() []int {
	var pages []int
	pageNums := p.GetPagesCnt()
	page := p.GetCurPage()

	switch {
	case page >= pageNums-4 && pageNums > 9:
		start := pageNums - 9 + 1
		pages = make([]int, 9)
		for i := range pages {
			pages[i] = start + i
		}
	case page >= 5 && pageNums > 9:
		start := page - 5 + 1
		pages = make([]int, int(math.Min(9, float64(page+4+1))))
		for i := range pages {
			pages[i] = start + i
		}
	default:
		pages = make([]int, int(math.Min(9, float64(pageNums))))
		for i := range pages {
			pages[i] = i + 1
		}
	}

	return pages
}

func (p *Pager) GetLink(pageNum int, pageSize int) string {
	link, _ := url.ParseRequestURI(p.req.URL.String())
	values := link.Query()
	total := p.GetPagesCnt()

	if pageNum <= 1 || pageNum > total {
		values.Del(p.curPageKey)
	} else {
		values.Set(p.curPageKey, strconv.Itoa(pageNum))
	}

	if pageSize > 0 {
		values.Set(p.pageSizeKey, strconv.Itoa(pageSize))
	}

	link.RawQuery = values.Encode()

	return link.String()
}

func (p *Pager) GetLinkPrev() string {
	return p.GetLink(p.GetCurPage()-1, 0)
}

func (p *Pager) GetLinkNext() string {
	return p.GetLink(p.GetCurPage()+1, 0)
}

func (p *Pager) GetLinkFirst() string {
	return p.GetLink(1, 0)
}

func (p *Pager) GetLinkLast() string {
	return p.GetLink(p.GetPagesCnt(), 0)
}
