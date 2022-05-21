package interfaces

import (
	"io"
)

// Pool is a pool interface.
type Pool interface {
	io.Closer
	AddWorker(Worker) error
	AddTask(Task) error
}
