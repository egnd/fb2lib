package tasks

import (
	"bytes"
	"fmt"
	"io"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/pkg/errors"
)

type PushReadTask func(io.ReadCloser, entities.BookInfo) error

type ReadTask struct {
	id          string
	reader      io.ReadCloser
	doParseTask PushParseTask
}

func NewReadTask(
	src string,
	lib string,
	reader io.ReadCloser,
	doParseTask PushParseTask,
) *ReadTask {
	return &ReadTask{
		id:          fmt.Sprintf("read [%s] %s", lib, src),
		reader:      reader,
		doParseTask: doParseTask,
	}
}

func (t *ReadTask) ID() string {
	return t.id
}

func (t *ReadTask) Do() error {
	defer t.reader.Close()

	data, err := io.ReadAll(t.reader)
	if err != nil {
		return errors.Wrap(err, "read data error")
	}

	if err = t.doParseTask(bytes.NewBuffer(data)); err != nil {
		return errors.Wrap(err, "do parse error")
	}

	return nil
}
