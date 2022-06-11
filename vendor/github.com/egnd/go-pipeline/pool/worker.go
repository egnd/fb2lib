package pool

import (
	"errors"
	"fmt"

	"github.com/egnd/go-pipeline"
)

// Worker struct for handling tasks.
type Worker struct {
	tasks chan pipeline.Task
	stop  chan struct{}
}

// NewWorker creates workers for pool pipeline.
func NewWorker(notifier chan<- pipeline.Doer, middlewares ...pipeline.DoerDecorator) *Worker {
	worker := &Worker{
		tasks: make(chan pipeline.Task, 1),
		stop:  make(chan struct{}),
	}

	go func() {
		execute := pipeline.DecorateDoer(func(task pipeline.Task) error {
			return task.Do() //nolint:wrapcheck
		}, middlewares...)

		for {
			if err := worker.notify(notifier); err != nil {
				return
			}

			for task := range worker.tasks {
				_ = execute(task)

				break
			}
		}
	}()

	return worker
}

func (w *Worker) notify(notifier chan<- pipeline.Doer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = w.Close()
		}
	}()

	select {
	case notifier <- w:
	case <-w.stop:
		err = errors.New("worker is stopped")
	}

	return
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
	w.stop <- struct{}{}
	close(w.stop)

	return
}
