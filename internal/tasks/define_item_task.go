package tasks

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/pkg/errors"
)

type DefineItemTask struct {
	id        string
	item      string
	lib       entities.Library
	repoMarks entities.ILibMarksRepo
	doFB2Task PushReadTask
	doZIPTask DoReadZipTask
}

func NewDefineItemTask(
	item string,
	lib entities.Library,
	repoMarks entities.ILibMarksRepo,
	doFB2Task PushReadTask,
	doZIPTask DoReadZipTask,
) *DefineItemTask {
	return &DefineItemTask{
		id:        fmt.Sprintf("look at [%s] %s", lib.Name, strings.TrimPrefix(item, lib.Dir)),
		item:      item,
		lib:       lib,
		repoMarks: repoMarks,
		doFB2Task: doFB2Task,
		doZIPTask: doZIPTask,
	}
}

func (t *DefineItemTask) ID() string {
	return t.id
}

func (t *DefineItemTask) Do() error {
	finfo, err := os.Stat(t.item)
	if err != nil {
		return errors.Wrap(err, "stat error")
	}

	if t.repoMarks.MarkExists(t.item) {
		return fmt.Errorf("%s already indexed", t.item)
	}

	switch path.Ext(t.item) {
	case ".zip":
		if err := t.doZIPTask(finfo); err != nil {
			return errors.Wrap(err, "do zip error")
		}

		if err := t.repoMarks.AddMark(t.item); err != nil {
			return errors.Wrap(err, "memorize zip error")
		}
	case ".fb2":
		reader, err := os.Open(t.item)
		if err != nil {
			return errors.Wrap(err, "open fb2 error")
		}

		if err := t.doFB2Task(reader, entities.BookInfo{
			LibName: t.lib.Name,
			Size:    uint64(finfo.Size()),
			Src:     strings.TrimPrefix(t.item, t.lib.Dir),
		}); err != nil {
			return errors.Wrap(err, "do fb2 error")
		}
	default:
		return fmt.Errorf("type error: unhandled type %s", path.Ext(t.item))
	}

	return nil
}
