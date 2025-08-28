package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	headersHit = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_headers_hit_total",
			Help: "Number of cache hits (http headers)",
		},
	)
	headersMiss = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_headers_miss_total",
			Help: "Number of cache miss (http headers)",
		},
	)
	memHit = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_mem_hit_total",
			Help: "Number of cache hits (server level)",
		},
	)
	memMiss = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_mem_miss_total",
			Help: "Number of cache miss (server level)",
		},
	)
	reset = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "reset",
			Help: "Number of reset",
		},
	)
)

func init() {
	prometheus.MustRegister(headersHit, memHit, reset)
}

var Cache metricsCache

type metricsCache struct{}

func (c *metricsCache) HeadersHit()  { headersHit.Inc() }
func (c *metricsCache) HeadersMiss() { headersMiss.Inc() }
func (c *metricsCache) MemHit()      { memHit.Inc() }
func (c *metricsCache) MemMiss()     { memMiss.Inc() }
func (c *metricsCache) ResetAll()    { reset.Inc() }
