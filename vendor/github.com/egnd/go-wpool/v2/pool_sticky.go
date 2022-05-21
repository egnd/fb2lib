package wpool

import (
	"errors"
	"sync"

	"github.com/cespare/xxhash"
	"github.com/egnd/go-wpool/v2/interfaces"
)

// StickyPool is a pool of "sticky" workers.
type StickyPool struct {
	tasks     chan interfaces.Task
	workers   []interfaces.Worker
	mxWorkers sync.Mutex
	logger    interfaces.Logger
}

// NewStickyPool creates pool of "sticky" workers.
func NewStickyPool(logger interfaces.Logger) *StickyPool {
	pool := &StickyPool{ //nolint:exhaustivestruct
		tasks:  make(chan interfaces.Task),
		logger: logger,
	}

	go pool.run()

	return pool
}

func (p *StickyPool) run() {
	for task := range p.tasks {
		worker, num := p.getWorkerFor(task)

		if err := worker.Do(task); err != nil {
			p.logger.Errorf(err, `worker #%d doing task "%s"`, num, task.GetID())
		}
	}
}

func (p *StickyPool) getWorkerFor(task interfaces.Task) (interfaces.Worker, uint64) { //nolint:ireturn
	p.mxWorkers.Lock()
	defer p.mxWorkers.Unlock()

	workerNum := xxhash.Sum64String(task.GetID()) % uint64(len(p.workers))

	return p.workers[workerNum], workerNum
}

// AddWorker is registering worker in the pool.
func (p *StickyPool) AddWorker(worker interfaces.Worker) (err error) {
	p.mxWorkers.Lock()
	defer p.mxWorkers.Unlock()

	p.workers = append(p.workers, worker)

	return
}

// AddTask is putting task into pool.
func (p *StickyPool) AddTask(task interfaces.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("pool is closed")
		}
	}()

	p.tasks <- task

	return
}

// Close is stopping pool and workers.
func (p *StickyPool) Close() (err error) {
	close(p.tasks)

	p.mxWorkers.Lock()
	defer p.mxWorkers.Unlock()

	for _, worker := range p.workers {
		if err = worker.Close(); err != nil {
			return
		}
	}

	return
}
