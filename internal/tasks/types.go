package tasks

import (
	"io"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-pipeline"
	"github.com/rs/zerolog"
)

type IndexTaskFactory func(io.Reader, entities.BookInfo, zerolog.Logger) pipeline.Task
