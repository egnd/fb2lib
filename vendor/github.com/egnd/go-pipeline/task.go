package pipeline

// TaskDecorator is a decorator for task execution logic.
type TaskDecorator func(TaskExecutor) TaskExecutor

// TaskExecutor is a type for task execution method.
type TaskExecutor func(Task) error

// Task is a task interface.
type Task interface {
	ID() string
	Do() error
}

// NewTaskExecutor builds chain of decorators.
func NewTaskExecutor(decorators []TaskDecorator) TaskExecutor {
	res := func(task Task) error {
		return task.Do() //nolint:wrapcheck
	}

	if len(decorators) == 0 {
		return res
	}

	for i := len(decorators) - 1; i >= 0; i-- {
		res = decorators[i](res)
	}

	return res
}
