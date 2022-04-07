// Package wpool contains structs and functions for making a pool of workers.
package wpool

import "errors"

// PoolCfg is pool config.
type PoolCfg struct {
	TasksBufSize uint
	WorkersCnt   uint
}

// Pool is struct for handlindg tasks with workers.
type Pool struct {
	pipeline chan IWorker
	tasks    chan ITask
	workers  []IWorker
}

// NewPool is a factory method for pool of workers.
func NewPool(cfg PoolCfg, newWorker IWorkerFactory) *Pool {
	pool := &Pool{
		tasks:    make(chan ITask, cfg.TasksBufSize),
		pipeline: make(chan IWorker),
		workers:  make([]IWorker, 0, cfg.WorkersCnt),
	}

	for i := uint(0); i < cfg.WorkersCnt; i++ {
		pool.workers = append(pool.workers, newWorker(i, pool.pipeline))
	}

	go func() {
		for worker := range pool.pipeline {
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

// Add is method for putting task into pool of workers.
func (p *Pool) Add(task ITask) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("add task to pool error: pool is closed")
		}
	}()

	p.tasks <- task

	return
}

// Close is method for stopping for pool of workers.
func (p *Pool) Close() (err error) {
	close(p.tasks)
	defer close(p.pipeline)

	for _, worker := range p.workers {
		if err = worker.Close(); err != nil {
			return
		}
	}

	return
}
