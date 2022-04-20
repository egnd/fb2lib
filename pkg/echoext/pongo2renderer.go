package echoext

import (
	"errors"
	"io"

	pongo2 "github.com/flosch/pongo2/v5"
	"github.com/labstack/echo/v4"
)

type PongoRendererCfg struct {
	Debug   bool
	TplsDir string
}

type PongoRenderer struct {
	set *pongo2.TemplateSet
}

func NewPongoRenderer(
	cfg PongoRendererCfg,
	data map[string]interface{},
	filters map[string]pongo2.FilterFunction,
) (*PongoRenderer, error) {
	for name, filter := range filters {
		pongo2.RegisterFilter(name, filter)
	}

	loader, err := pongo2.NewLocalFileSystemLoader(cfg.TplsDir)
	if err != nil {
		return nil, err
	}

	res := PongoRenderer{
		set: pongo2.NewSet("echo_renderer", loader),
	}

	res.set.Debug = cfg.Debug
	for k, v := range data {
		res.set.Globals[k] = v
	}

	return &res, nil
}

func (r *PongoRenderer) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	tpl, err := r.set.FromCache(name)
	if err != nil {
		return err
	}

	var pongoCtx pongo2.Context
	if data != nil {
		var ok bool
		if pongoCtx, ok = data.(pongo2.Context); !ok {
			return errors.New("pongo renderer error: invalid data type")
		}
	}

	return tpl.ExecuteWriter(pongoCtx, w)
}
