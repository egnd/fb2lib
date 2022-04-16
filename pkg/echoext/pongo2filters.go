package echoext

import (
	"strings"

	"github.com/dustin/go-humanize"
	pongo2 "github.com/flosch/pongo2/v5"
)

func PongoFilterFileSize(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(humanize.IBytes(uint64(in.Float()))), nil
}

func PongoFilterTrimSpace(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	var val string

	if !in.IsNil() {
		val = strings.TrimSpace(in.String())
	}

	return pongo2.AsValue(val), nil
}
