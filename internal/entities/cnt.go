package entities

import (
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type CntAtomic32 uint32

func (c *CntAtomic32) Inc(delta uint32) uint32 {
	return atomic.AddUint32((*uint32)(c), delta)
}

func (c *CntAtomic32) Total() uint32 {
	return atomic.LoadUint32((*uint32)(c))
}

type PromCounter struct {
	metric *prometheus.CounterVec
	pusher *push.Pusher
}

func NewPromCounter(metric, job string, params ...string) *PromCounter {
	res := &PromCounter{
		metric: prometheus.NewCounterVec(prometheus.CounterOpts{Name: metric}, params),
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(res.metric)
	res.pusher = push.New("http://push.prom.rpi", job).
		Gatherer(registry).Client(http.DefaultClient)

	return res
}

func (m *PromCounter) Push() error {
	return m.pusher.Add()
	return nil
}

func (m *PromCounter) Inc(values ...string) error {
	m.metric.WithLabelValues(values...).Inc()
	// return nil

	return m.Push()
}
