// Package wpool contains structs and functions for making a pool of workers.
package wpool

import (
	"context"
	"io"
)

// IPool is pool of workers interface.
type IPool interface {
	io.Closer
	Add(ITask) error
}

// IWorker is a worker interface.
type IWorker interface {
	io.Closer
	Do(ITask) error
}

// IWorkerFactory is a factory method to pass into pool of workers.
type IWorkerFactory func(num uint, pipeline chan IWorker) IWorker

// ITask is a task interface.
type ITask interface {
	Do(context.Context)
}
