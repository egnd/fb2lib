package wpool

import (
	"errors"

	"github.com/egnd/go-wpool/v2/interfaces"
)

// Worker struct for handling tasks.
type Worker struct {
	tasks chan interfaces.Task
}

// NewWorker creates workers with tasks queue.
func NewWorker(buffSize int) *Worker {
	worker := &Worker{
		tasks: make(chan interfaces.Task, buffSize),
	}

	go worker.run()

	return worker
}

func (w *Worker) run() {
	for task := range w.tasks {
		task.Do()
	}
}

// Do is putting task to worker queue.
func (w *Worker) Do(task interfaces.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("worker is closed")
		}
	}()

	w.tasks <- task

	return nil
}

// Close is stopping worker.
func (w *Worker) Close() error {
	close(w.tasks)

	return nil
}
