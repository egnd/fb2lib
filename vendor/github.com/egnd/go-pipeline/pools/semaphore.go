package pools

import (
	"fmt"
	"sync"

	"github.com/egnd/go-pipeline"
)

// Semaphore is a struct for tasks parallel execution.
type Semaphore struct {
	wg      *sync.WaitGroup
	limiter chan struct{}
	execute pipeline.TaskExecutor
}

// NewSemaphore is a factory for Semaphore.
func NewSemaphore(threadsCnt int, wg *sync.WaitGroup, decorators ...pipeline.TaskDecorator) *Semaphore {
	if wg == nil {
		wg = &sync.WaitGroup{}
	}

	return &Semaphore{
		wg:      wg,
		limiter: make(chan struct{}, threadsCnt),
		execute: pipeline.NewTaskExecutor(decorators),
	}
}

// Push is pushing task into semaphore.
func (p *Semaphore) Push(task pipeline.Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("semaphore do err: %v", r)
		}
	}()

	p.wg.Add(1)
	p.limiter <- struct{}{}

	go func() {
		defer func() {
			<-p.limiter
			p.wg.Done()
		}()

		_ = p.execute(task)
	}()

	return
}

// Wait blocks until tasks are completed.
func (p *Semaphore) Wait() {
	p.wg.Wait()
}

// Close is stopping Semaphore.
func (p *Semaphore) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("semaphore close err: %v", r)
		}
	}()

	close(p.limiter)

	return
}
