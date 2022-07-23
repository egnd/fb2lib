package workers

import (
	"fmt"
	"sync"

	"github.com/egnd/go-pipeline"
)

// Worker struct for handling tasks.
type Worker struct {
	tasks chan pipeline.Task
}

// NewWorker creates workers with tasks queue.
func NewWorker(queueSize int, wg *sync.WaitGroup, execute pipeline.TaskExecutor) *Worker {
	if wg == nil {
		panic("worker requires WaitGroup")
	}

	worker := &Worker{
		tasks: make(chan pipeline.Task, queueSize),
	}

	go func() {
		for task := range worker.tasks {
			_ = execute(task)

			wg.Done()
		}
	}()

	return worker
}

// Do is putting task to worker's queue.
func (w *Worker) Do(task pipeline.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("worker do err: %v", r)
		}
	}()

	w.tasks <- task

	return
}

// Close is stopping a worker.
func (w *Worker) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("worker close err: %v", r)
		}
	}()

	close(w.tasks)

	return
}
