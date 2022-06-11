// Package pool contains pool and worker structs
package pool

import (
	"fmt"

	"github.com/egnd/go-pipeline"
)

// Pool is a pool of workers.
type Pool struct {
	workers []pipeline.Doer
	tasks   chan pipeline.Task
}

// NewPool creates a pool of workers.
func NewPool(bus chan pipeline.Doer, workers ...pipeline.Doer) *Pool {
	pool := &Pool{
		workers: workers,
		tasks:   make(chan pipeline.Task),
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
func (p *Pool) Push(task pipeline.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pool push err: %v", r)
		}
	}()

	p.tasks <- task

	return
}

// Close is stopping pool and workers.
func (p *Pool) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pool close err: %v", r)
		}
	}()

	close(p.tasks)

	for _, worker := range p.workers {
		if err = worker.Close(); err != nil {
			return
		}
	}

	return
}
