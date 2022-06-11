package pipeline

import (
	"io"
)

// Doer is a worker interface.
type Doer interface {
	io.Closer
	Do(Task) error
}

// Tasker is a type for task execution method.
type Tasker func(Task) error

// DoerDecorator is a decorator for task execution logic.
type DoerDecorator func(Tasker) Tasker

// DecorateDoer builds chain of decorators.
func DecorateDoer(handler Tasker, middlewares ...DoerDecorator) Tasker {
	if len(middlewares) == 0 {
		return handler
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
