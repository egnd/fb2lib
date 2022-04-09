package pagination

type IPager interface {
	SetCurPage(val int, key ...string) IPager
	ReadCurPage(key ...string) IPager
	GetCurPage() int
	IsCurPage(val int) bool
	ReadPageSize(key ...string) IPager
	SetPageSize(val int, key ...string) IPager
	GetPageSize() int
	GetOffset() int
	SetTotal(val interface{}) IPager
	GetTotal() uint64
	GetPagesCnt() int
	HasPages() bool
	HasPrev() bool
	HasNext() bool
	GetPages() []int
	GetLink(pageNum, pageSize int) string
	GetLinkPrev() string
	GetLinkNext() string
	GetLinkFirst() string
	GetLinkLast() string
}
