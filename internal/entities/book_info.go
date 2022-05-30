package entities

type BookInfo struct {
	Offset         uint64    `json:"from"`
	Size           uint64    `json:"size"`
	SizeCompressed uint64    `json:"sizec"`
	LibName        string    `json:"lib"`
	Src            string    `json:"src"`
	Index          BookIndex `json:"-"`
}
