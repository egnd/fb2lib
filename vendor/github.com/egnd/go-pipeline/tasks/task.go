// Package tasks contains different types of tasks
package tasks

// FuncTask is a wrapper for task callback.
type FuncTask struct {
	id   string
	task func() error
}

// NewFunc is a factory for SomeTask.
func NewFunc(id string, task func() error) *FuncTask {
	return &FuncTask{
		id:   id,
		task: task,
	}
}

// ID returns task id.
func (t *FuncTask) ID() string {
	return t.id
}

// Do runs task.
func (t *FuncTask) Do() error {
	if t.task == nil {
		return nil
	}

	return t.task()
}
