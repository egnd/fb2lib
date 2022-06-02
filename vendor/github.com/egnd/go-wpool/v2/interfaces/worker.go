package interfaces

import (
	"io"
)

// Worker is a worker interface.
type Worker interface {
	io.Closer
	Do(Task) error
}
