// Package pools contains pool and worker structs
package pools

import (
	"fmt"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/workers"
)

// BusPool is a pool of workers.
type BusPool struct {
	doers []pipeline.Doer
	tasks chan pipeline.Task
}

// NewBusPool creates a pool of workers.
func NewBusPool(threadsCnt, queueSize int, decorators ...pipeline.TaskDecorator) *BusPool {
	if threadsCnt < 1 {
		panic("BusPool requires at least 1 thread")
	}

	bus := make(chan pipeline.Doer)
	executor := pipeline.NewTaskExecutor(decorators)
	pool := &BusPool{
		doers: make([]pipeline.Doer, threadsCnt),
		tasks: make(chan pipeline.Task, queueSize),
	}

	for k := range pool.doers {
		pool.doers[k] = workers.NewBusWorker(bus, executor)
	}

	go func() {
		for worker := range bus {
			for task := range pool.tasks {
				if err := worker.Do(task); err != nil {
					panic(err)
				}

				break //nolint:staticcheck
			}
		}
	}()

	return pool
}

// Push is pushing task into pool.
func (p *BusPool) Push(task pipeline.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pool push err: %v", r)
		}
	}()

	p.tasks <- task

	return
}

// Close is stopping pool and workers.
func (p *BusPool) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pool close err: %v", r)
		}
	}()

	close(p.tasks)

	for _, doer := range p.doers {
		if err = doer.Close(); err != nil {
			return
		}
	}

	return
}
