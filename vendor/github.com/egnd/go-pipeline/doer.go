package pipeline

import (
	"io"
)

// Doer is a worker interface.
type Doer interface {
	io.Closer
	Do(Task) error
}
