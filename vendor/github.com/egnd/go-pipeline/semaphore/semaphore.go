// Package semaphore contains semaphore struct
package semaphore

import (
	"fmt"
	"sync"

	"github.com/egnd/go-pipeline"
)

// Semaphore is a struct for tasks parallel execution.
type Semaphore struct {
	wg      sync.WaitGroup
	limiter chan struct{}
	execute pipeline.Tasker
}

// NewSemaphore is a factory for Semaphore.
func NewSemaphore(threadsCnt int, middlewares ...pipeline.DoerDecorator) *Semaphore {
	return &Semaphore{ //nolint:exhaustivestruct
		limiter: make(chan struct{}, threadsCnt),
		execute: pipeline.DecorateDoer(func(task pipeline.Task) error {
			return task.Do() //nolint:wrapcheck
		}, middlewares...),
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

// Close is stopping Semaphore.
func (p *Semaphore) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("semaphore close err: %v", r)
		}
	}()

	p.wg.Wait()
	close(p.limiter)

	return
}
