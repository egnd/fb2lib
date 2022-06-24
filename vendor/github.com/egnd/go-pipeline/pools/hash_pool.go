package pools

import (
	"fmt"
	"sync"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/workers"
)

// HashPool is a pool of "sticky" workers.
type HashPool struct {
	wg    *sync.WaitGroup
	tasks chan pipeline.Task
	doers []pipeline.Doer
}

// NewHashPool creates pool of "sticky" workers.
func NewHashPool(threadsCnt, queueSize int,
	wg *sync.WaitGroup, hasher pipeline.Hasher, decorators ...pipeline.TaskDecorator,
) *HashPool {
	if threadsCnt < 1 {
		panic("HashPool requires at least 1 thread")
	}

	if wg == nil {
		wg = &sync.WaitGroup{}
	}

	executor := pipeline.NewTaskExecutor(decorators)
	pool := &HashPool{
		wg:    wg,
		doers: make([]pipeline.Doer, threadsCnt),
		tasks: make(chan pipeline.Task, queueSize),
	}

	for k := range pool.doers {
		pool.doers[k] = workers.NewWorker(0, pool.wg, executor)
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

	p.wg.Add(1)
	p.tasks <- task

	return
}

// Wait blocks until tasks are completed.
func (p *HashPool) Wait() {
	p.wg.Wait()
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
