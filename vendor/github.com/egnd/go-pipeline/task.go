package pipeline

// Task is a task interface.
type Task interface {
	ID() string
	Do() error
}
