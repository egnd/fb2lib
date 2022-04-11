package render2echo

import (
	"github.com/dustin/go-humanize"
	pongo2 "github.com/flosch/pongo2/v5"
)

func FilterFileSize(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(humanize.Bytes(uint64(in.Float()))), nil
}
