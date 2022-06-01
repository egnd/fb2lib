package tasks

import (
	"io"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/egnd/go-wpool/v2/interfaces"
	"github.com/rs/zerolog"
)

type IndexTaskFactory func(io.Reader, entities.BookInfo, zerolog.Logger) interfaces.Task
