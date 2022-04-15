package echoext

import (
	"github.com/dustin/go-humanize"
	pongo2 "github.com/flosch/pongo2/v5"
)

func PongoFilterFileSize(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(humanize.IBytes(uint64(in.Float()))), nil
}