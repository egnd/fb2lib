package entities

import (
	"regexp"
	"strings"
)

var sanitizeFreqKey = regexp.MustCompile(`[^a-zа-я0-9]+`)

type ItemFreq struct {
	Val  string `json:"val,omitempty"`
	Freq int    `json:"frq,omitempty"`
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

type ItemFreqMap map[string]ItemFreq

func (m ItemFreqMap) Put(val string, freq int) {
	val = strings.TrimSpace(strings.Split(val, "(")[0])
	key := sanitizeFreqKey.ReplaceAllString(strings.ToLower(val), "")

	if key == "" {
		return
	}

	if old, ok := m[key]; ok {
		freq += old.Freq
	}

	m[key] = ItemFreq{Val: val, Freq: freq}
}
