package entities

type ItemFreq struct {
	Val  string
	Freq int
}

type FreqsItems []ItemFreq

func (g FreqsItems) Len() int {
	return len(g)
}

func (g FreqsItems) Less(i, j int) bool {
	return g[i].Freq < g[j].Freq
}

func (g FreqsItems) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}
