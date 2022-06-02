# go-wpool

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/go-wpool.svg)](https://pkg.go.dev/github.com/egnd/go-wpool)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/go-wpool)](https://goreportcard.com/report/github.com/egnd/go-wpool)
[![Coverage](https://gocover.io/_badge/github.com/egnd/go-wpool)](https://gocover.io/github.com/egnd/go-wpool)
[![Pipeline](https://github.com/egnd/go-wpool/actions/workflows/pipeline.yml/badge.svg)](https://github.com/egnd/go-wpool/actions?query=workflow%3APipeline)

Golang package for making a pool of workers.

### Pool example:
```golang
package main

import (
	"fmt"
	"sync"

	"github.com/egnd/go-wpool/v2"
	"github.com/egnd/go-wpool/v2/interfaces"
	"github.com/rs/zerolog"
)

func main() {
	// create pipeline and pool
	pipeline := make(chan interfaces.Worker)
	pool := wpool.NewPipelinePool(pipeline, 
		wpool.NewZerologAdapter(zerolog.New()),
	)
	defer pool.Close()

	// add few workers
	pool.AddWorker(wpool.NewPipelineWorker(pipeline))
	pool.AddWorker(wpool.NewPipelineWorker(pipeline))

	// put some tasks to pool
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		if err := pool.AddTask(&SomeTask{&wg, "task"+fmt.Sprint(i)}); err != nil {
			panic(err)
		}
	}

	// wait for tasks to be completed
	wg.Wait()
}
```

### Sticky pool example (tasks with the same ID will be processed by the same worker):
```golang
package main

import (
	"fmt"
	"sync"

	"github.com/egnd/go-wpool/v2"
	"github.com/egnd/go-wpool/v2/interfaces"
	"github.com/rs/zerolog"
)

func main() {
	// create pool
	pool := wpool.NewStickyPool(
		wpool.NewZerologAdapter(zerolog.Nop())
	)
	defer pool.Close()

	// add few workers
	buffSize := 100
	pool.AddWorker(wpool.NewWorker(buffSize))
	pool.AddWorker(wpool.NewWorker(buffSize))

	// put some tasks to pool
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		if err := pool.AddTask(&SomeTask{&wg, "task"+fmt.Sprint(i)}); err != nil {
			panic(err)
		}
	}

	// wait for tasks to be completed
	wg.Wait()
}
```