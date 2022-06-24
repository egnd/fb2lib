// Package pools contains pool and worker structs
package pools

import (
	"fmt"
	"sync"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/workers"
)

// BusPool is a pool of workers.
type BusPool struct {
	wg    *sync.WaitGroup
	doers []pipeline.Doer
	tasks chan pipeline.Task
}

// NewBusPool creates a pool of workers.
func NewBusPool(threadsCnt, queueSize int, wg *sync.WaitGroup, decorators ...pipeline.TaskDecorator) *BusPool {
	if threadsCnt < 1 {
		panic("BusPool requires at least 1 thread")
	}

	if wg == nil {
		wg = &sync.WaitGroup{}
	}

	bus := make(chan pipeline.Doer)
	executor := pipeline.NewTaskExecutor(decorators)
	pool := &BusPool{
		wg:    wg,
		doers: make([]pipeline.Doer, threadsCnt),
		tasks: make(chan pipeline.Task, queueSize),
	}

	for k := range pool.doers {
		pool.doers[k] = workers.NewBusWorker(bus, pool.wg, executor)
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

	p.wg.Add(1)
	p.tasks <- task

	return
}

// Wait blocks until tasks are completed.
func (p *BusPool) Wait() {
	p.wg.Wait()
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
