// Package workers contains different types of workers
package workers

import (
	"errors"
	"fmt"
	"sync"

	"github.com/egnd/go-pipeline"
)

// BusWorker struct for handling tasks.
type BusWorker struct {
	tasks chan pipeline.Task
	stop  chan struct{}
}

// NewBusWorker creates workers for pool pipeline.
func NewBusWorker(bus chan<- pipeline.Doer, wg *sync.WaitGroup, execute pipeline.TaskExecutor) *BusWorker {
	if wg == nil {
		panic("worker requires WaitGroup")
	}

	worker := &BusWorker{
		tasks: make(chan pipeline.Task, 1),
		stop:  make(chan struct{}),
	}

	go func() {
		for {
			if err := worker.notify(bus); err != nil {
				return
			}

			for task := range worker.tasks {
				_ = execute(task)

				wg.Done()

				break
			}
		}
	}()

	return worker
}

func (w *BusWorker) notify(bus chan<- pipeline.Doer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = w.Close()
		}
	}()

	select {
	case bus <- w:
	case <-w.stop:
		err = errors.New("worker is stopped")
	}

	return
}

// Do is putting task to worker's queue.
func (w *BusWorker) Do(task pipeline.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("worker do err: %v", r)
		}
	}()

	w.tasks <- task

	return
}

// Close is stopping a worker.
func (w *BusWorker) Close() (err error) {
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
