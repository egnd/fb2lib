package entities

type GenreFreq struct {
	Name string
	Cnt  uint32
}

type GenresIndex []GenreFreq

func (g GenresIndex) Len() int {
	return len(g)
}

func (g GenresIndex) Less(i, j int) bool {
	return g[i].Cnt < g[j].Cnt
}

func (g GenresIndex) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
