// Package wpool contains structs and functions for making a pool of workers.
package wpool

import (
	"errors"
	"sync"

	"github.com/egnd/go-wpool/v2/interfaces"
)

// PipelinePool is a pool of workers with pipeline.
type PipelinePool struct {
	tasks     chan interfaces.Task
	workers   []interfaces.Worker
	mxWorkers sync.Mutex
	logger    interfaces.Logger
}

// NewPipelinePool creates pool of workers with pipeline.
func NewPipelinePool(pipeline chan interfaces.Worker, logger interfaces.Logger) *PipelinePool {
	pool := &PipelinePool{ //nolint:exhaustivestruct
		tasks:  make(chan interfaces.Task),
		logger: logger,
	}

	go pool.run(pipeline)

	return pool
}

func (p *PipelinePool) run(pipeline chan interfaces.Worker) {
	for worker := range pipeline {
		for task := range p.tasks {
			if err := worker.Do(task); err != nil {
				p.logger.Errorf(err, `passing task "%s" to worker`, task.GetID())
			}

			break //nolint:staticcheck
		}
	}
}

// AddWorker is registering worker in the pool.
func (p *PipelinePool) AddWorker(worker interfaces.Worker) (err error) {
	p.mxWorkers.Lock()
	defer p.mxWorkers.Unlock()

	p.workers = append(p.workers, worker)

	return
}

// AddTask is putting task into pool.
func (p *PipelinePool) AddTask(task interfaces.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("pool is closed")
		}
	}()

	p.tasks <- task

	return
}

// Close is stopping pool and workers.
func (p *PipelinePool) Close() (err error) {
	close(p.tasks)

	p.mxWorkers.Lock()
	defer p.mxWorkers.Unlock()

	for _, worker := range p.workers {
		if wErr := worker.Close(); wErr != nil {
			err = wErr
		}
	}

	return
}
