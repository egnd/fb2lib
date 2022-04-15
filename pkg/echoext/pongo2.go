package echoext

import (
	"errors"
	"io"

	pongo2 "github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

type PongoRenderer struct {
	debug bool
	data  map[string]interface{}
}

func NewPongoRenderer(
	debug bool,
	data map[string]interface{},
	filters map[string]pongo2.FilterFunction,
) *PongoRenderer {
	for name, filter := range filters {
		pongo2.RegisterFilter(name, filter)
	}

	return &PongoRenderer{
		debug: debug,
		data:  data,
	}
}

func (r *PongoRenderer) Render(w io.Writer, name string, data interface{}, ctx echo.Context) (err error) {
	var pongoCtx pongo2.Context
	var tpl *pongo2.Template

	if data != nil {
		var ok bool
		if pongoCtx, ok = data.(pongo2.Context); !ok {
			return errors.New("no pongo2.Context data was passed")
		}
	}

	if r.debug {
		tpl, err = pongo2.FromFile(name)
	} else {
		tpl, err = pongo2.FromCache(name)
	}

	if err != nil {
		return err
	}

	for k, v := range r.data {
		pongoCtx[k] = v
	}

	return tpl.ExecuteWriter(pongoCtx, w)
}
