package wpool

import (
	"context"
)

// Task is a task struct.
type Task struct {
	Callback func(context.Context)
}

// Do is executing task logic.
func (t *Task) Do(ctx context.Context) {
	if t.Callback == nil {
		return
	}

	t.Callback(ctx)
}
