# go-wpool

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/go-wpool.svg)](https://pkg.go.dev/github.com/egnd/go-wpool)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/go-wpool)](https://goreportcard.com/report/github.com/egnd/go-wpool)
[![Coverage](https://gocover.io/_badge/github.com/egnd/go-wpool)](https://gocover.io/github.com/egnd/go-wpool)
[![Pipeline](https://github.com/egnd/go-wpool/actions/workflows/pipeline.yml/badge.svg)](https://github.com/egnd/go-wpool/actions?query=workflow%3APipeline)

Golang package for making a pool of workers.

### Pool example:
```golang
ctx := context.Background()

// create pool and define worker's factory
pool := wpool.NewPool(wpool.PoolCfg{
    WorkersCnt:   3,
    TasksBufSize: 10,
}, func(num uint, pipeline chan wpool.IWorker) wpool.IWorker {
    return wpool.NewWorker(ctx, wpool.WorkerCfg{
        Pipeline: pipeline,
        TaskTTL:  300 * time.Millisecond,
    })
})
defer pool.Close()

// put some tasks to pool
var wg sync.WaitGroup
for i := 0; i < 20; i++ {
    wg.Add(1)
    if err := pool.Add(&wpool.Task{Callback: func(tCtx context.Context, task *wpool.Task) {
        // @TODO: do something here
        return
    }}); err != nil {
        panic(err)
    }
}
// wait for tasks to be completed
wg.Wait()
```

### Worker example:
```golang
ctx := context.Background()

// create worker
worker := wpool.NewWorker(ctx, wpool.WorkerCfg{
    TasksChanBuff: 10,
    TaskTTL:       300 * time.Duration,
})
defer worker.Close()

// put some tasks to worker
var wg sync.WaitGroup
for i := 0; i < 20; i++ {
    wg.Add(1)
    if err := worker.Do(&wpool.Task{Callback: func(tCtx context.Context, task *wpool.Task) error {
        // @TODO: do something here
        return
    }}); err != nil {
        panic(err)
    }
}
// wait for tasks to be completed
wg.Wait()
```
