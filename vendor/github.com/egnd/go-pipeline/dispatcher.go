// Package pipeline contains tools for parallel execution
package pipeline

import (
	"io"
)

// Dispatcher is a pool interface.
type Dispatcher interface {
	io.Closer
	Push(Task) error
}
