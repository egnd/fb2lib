package wpool

import (
	"context"
	"errors"
	"time"
)

// WorkerCfg is a config for Worker.
type WorkerCfg struct {
	TasksChanBuff uint
	TaskTTL       time.Duration
	Pipeline      chan<- IWorker
}

// Worker is a struct for handling tasks.
type Worker struct {
	cfg        WorkerCfg
	tasks      chan ITask
	stopNotify chan struct{}
}

// NewWorker is a factory method for creating of new workers.
func NewWorker(ctx context.Context, cfg WorkerCfg) *Worker {
	tasksBuf := cfg.TasksChanBuff

	if cfg.Pipeline != nil {
		tasksBuf = 1
	}

	worker := &Worker{
		cfg:        cfg,
		tasks:      make(chan ITask, tasksBuf),
		stopNotify: make(chan struct{}),
	}

	go worker.run(ctx)

	return worker
}

func (w *Worker) run(ctx context.Context) {
	if w.cfg.Pipeline == nil {
		for task := range w.tasks {
			w.exec(ctx, task)
		}

		return
	}

	for {
		if err := w.notifyPipeline(); err != nil {
			return
		}

		for task := range w.tasks {
			w.exec(ctx, task)

			break
		}
	}
}

func (w *Worker) notifyPipeline() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("move worker to pipeline error: pipeline is closed")
		}
	}()

	select {
	case w.cfg.Pipeline <- w:
	case <-w.stopNotify:
		err = errors.New("worker is stopped")
	}

	return
}

func (w *Worker) exec(ctx context.Context, task ITask) {
	if w.cfg.TaskTTL > 0 {
		var ctxCancel context.CancelFunc
		ctx, ctxCancel = context.WithTimeout(ctx, w.cfg.TaskTTL)

		defer ctxCancel()
	}

	task.Do(ctx)
}

// Do is method for putting task to worker.
func (w *Worker) Do(task ITask) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("add task to worker error: worker is closed")
		}
	}()

	w.tasks <- task

	return nil
}

// Close is a method for worker stopping.
func (w *Worker) Close() (err error) {
	close(w.tasks)

	if w.cfg.Pipeline != nil {
		w.stopNotify <- struct{}{}
	}

	close(w.stopNotify)

	return
}
