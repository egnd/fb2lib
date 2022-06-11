# go-pipeline

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/go-pipeline.svg)](https://pkg.go.dev/github.com/egnd/go-pipeline)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/go-pipeline)](https://goreportcard.com/report/github.com/egnd/go-pipeline)
[![Coverage](https://gocover.io/_badge/github.com/egnd/go-pipeline)](https://gocover.io/github.com/egnd/go-pipeline)
[![Pipeline](https://github.com/egnd/go-pipeline/actions/workflows/pipeline.yml/badge.svg)](https://github.com/egnd/go-pipeline/actions?query=workflow%3APipeline)

Golang package for parallel execution of tasks.

### Pool:
Common Pool of Workers. The Task is taken into work by the first released Worker.
```golang
package main

import (
	"log"
	"sync"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/pool"
)

func main() {
	// create notifications channel
	notifier := make(chan pipeline.Doer)

	// create pool
	pipe := pool.NewPool(notifier,
		// create and register workers
		pool.NewWorker(notifier),
		// also it is possible to add some middlewares at tasks execution
		pool.NewWorker(notifier, func(next pipeline.Tasker) pipeline.Tasker {
			return func(task pipeline.Task) error {
				if err := next(task); err != nil {
					log.Println(err)
				}
				return nil
			}
		}),
	)

	// start producing tasks to pool
	var wg sync.WaitGroup
	pipe.Push(NewTask(&wg))
	pipe.Push(NewTask(&wg))
	pipe.Push(NewTask(&wg))
	wg.Wait()

	// close pool
	if err := pipe.Close(); err != nil {
		panic(err)
	}
}

```

### Hashing pool:
A Pool of Workers, but Tasks with the same ID will be processed by the same worker.
```golang
package main

import (
	"log"
	"sync"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/hashpool"
)

func main() {
	// create pool
	pipe := hashpool.NewPool(
		// define task's ID hashing func
		hashpool.DefaultHasher,
		// create and register workers
		hashpool.NewWorker(queueSize),
		// also it is possible to add some middlewares at tasks execution
		hashpool.NewWorker(queueSize, func(next pipeline.Tasker) pipeline.Tasker {
			return func(task pipeline.Task) error {
				if err := next(task); err != nil {
					log.Println(err)
				}
				return nil
			}
		}),
	)

	// start producing tasks to pool
	var wg sync.WaitGroup
	pipe.Push(NewTask(&wg))
	pipe.Push(NewTask(&wg))
	pipe.Push(NewTask(&wg))
	wg.Wait()

	// close pool
	if err := pipe.Close(); err != nil {
		panic(err)
	}
}

```

### Semaphore:
Primitive for limiting the number of threads for the Tasks parallel execution.
```golang
package main

import (
	"log"
	"sync"

	"github.com/egnd/go-pipeline"
	"github.com/egnd/go-pipeline/semaphore"
)

func main() {
	// create semaphore
	pipe := semaphore.NewSemaphore(threadsCount,
		// it is possible to add some middlewares at tasks execution
		func(next pipeline.Tasker) pipeline.Tasker {
			return func(task pipeline.Task) error {
				if err := next(task); err != nil {
					log.Println(err)
				}
				return nil
			}
		},
	)

	// start producing tasks to pool
	var wg sync.WaitGroup
	pipe.Push(NewTask(&wg))
	pipe.Push(NewTask(&wg))
	pipe.Push(NewTask(&wg))
	wg.Wait()

	// close pool
	if err := pipe.Close(); err != nil {
		panic(err)
	}
}

```
