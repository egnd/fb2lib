package pools

import (
	"fmt"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/workers"
)

// HashPool is a pool of "sticky" workers.
type HashPool struct {
	tasks chan pipeline.Task
	doers []pipeline.Doer
}

// NewHashPool creates pool of "sticky" workers.
func NewHashPool(threadsCnt, queueSize int, hasher pipeline.Hasher, decorators ...pipeline.TaskDecorator) *HashPool {
	if threadsCnt < 1 {
		panic("HashPool requires at least 1 thread")
	}

	executor := pipeline.NewTaskExecutor(decorators)
	pool := &HashPool{
		doers: make([]pipeline.Doer, threadsCnt),
		tasks: make(chan pipeline.Task, queueSize),
	}

	for k := range pool.doers {
		pool.doers[k] = workers.NewWorker(0, executor)
	}

	go func() {
		for task := range pool.tasks {
			if err := pool.doers[hasher(task.ID(), uint64(threadsCnt))].Do(task); err != nil {
				panic(err)
			}
		}
	}()

	return pool
}

// Push is putting task into pool.
func (p *HashPool) Push(task pipeline.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pool push err: %v", r)
		}
	}()

	p.tasks <- task

	return
}

// Close is stopping pool and workers.
func (p *HashPool) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pool close err: %v", r)
		}
	}()

	close(p.tasks)

	for _, worker := range p.doers {
		if err = worker.Close(); err != nil {
			return
		}
	}

	return
}
