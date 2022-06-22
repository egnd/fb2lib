# go-pipeline

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/go-pipeline.svg)](https://pkg.go.dev/github.com/egnd/go-pipeline)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/go-pipeline)](https://goreportcard.com/report/github.com/egnd/go-pipeline)
[![Coverage](https://gocover.io/_badge/github.com/egnd/go-pipeline?k1)](https://gocover.io/github.com/egnd/go-pipeline)
[![Pipeline](https://github.com/egnd/go-pipeline/actions/workflows/pipeline.yml/badge.svg)](https://github.com/egnd/go-pipeline/actions?query=workflow%3APipeline)

Golang package for parallel execution of tasks.

### Pools types:
* BusPool: Common Pool of Workers. The Task is taken into work by the first released Worker.
* HashPool: Worker pool, which allows you to change the strategy for assigning Tasks to Workers.
* Semaphore: Primitive for limiting the number of threads for the Tasks parallel execution.

### Examples:
```golang
package main

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"go.uber.org/zap"

	"github.com/egnd/go-pipeline/assign"
	"github.com/egnd/go-pipeline/decorators"
	"github.com/egnd/go-pipeline/pools"
	"github.com/egnd/go-pipeline/tasks"
)

func main() {
	// BusPool example:
	pipe := pools.NewBusPool(
		2,  // set parallel threads count
		10, // set tasks queue size
		// add some task decorators:
		decorators.LogErrorZero(zerolog.Nop()), // log tasks errors with zerolog logger
		decorators.CatchPanic,                  // convert tasks panics to errors
	)

	// HashPool example:
	pipe := pools.NewHashPool(
		2,             // set parallel threads count
		10,            // set tasks queue size
		assign.Sticky, // choose tasks to workers assignment method
		// add some task decorators:
		decorators.LogErrorZap(zap.NewNop()), // log tasks errors with zap logger
		decorators.CatchPanic,                // convert tasks panics to errors
	)

	// Semaphore example:
	pipe := pools.NewSemaphore(2, // set parallel threads count
		// add some task decorators:
		decorators.ThrowPanic, // convert tasks errors to panics
	)

	// Send some tasks to pool
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		pipe.Push(tasks.NewFunc("task#"+fmt.Sprint(i), func() error {
			defer wg.Done()
			return nil
		}))
	}

	// Wait for task processing
	wg.Wait()

	// Close pool
	if err := pipe.Close(); err != nil {
		panic(err)
	}
}
```
