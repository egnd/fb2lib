package wpool

import (
	"errors"

	"github.com/egnd/go-wpool/v2/interfaces"
)

// PipelineWorker struct for handling tasks.
type PipelineWorker struct {
	tasks chan interfaces.Task
	stop  chan struct{}
}

// NewPipelineWorker creates workers for pool pipeline.
func NewPipelineWorker(pipeline chan<- interfaces.Worker) *PipelineWorker {
	worker := &PipelineWorker{
		tasks: make(chan interfaces.Task, 1),
		stop:  make(chan struct{}),
	}

	go worker.run(pipeline)

	return worker
}

func (w *PipelineWorker) run(pipeline chan<- interfaces.Worker) {
	for {
		if err := w.notifyPipeline(pipeline); err != nil {
			return
		}

		for task := range w.tasks {
			task.Do()

			break
		}
	}
}

func (w *PipelineWorker) notifyPipeline(pipeline chan<- interfaces.Worker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = w.Close()
		}
	}()

	select {
	case pipeline <- w:
	case <-w.stop:
		err = errors.New("worker is stopped")
	}

	return
}

// Do is putting task to worker queue.
func (w *PipelineWorker) Do(task interfaces.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("worker is closed")
		}
	}()

	w.tasks <- task

	return
}

// Close is stopping worker.
func (w *PipelineWorker) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("worker already closed")
		}
	}()

	close(w.tasks)
	w.stop <- struct{}{}
	close(w.stop)

	return
}
