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
	linodeAssetHit = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "linode_asset_hit_total",
			Help: "Number of successful asset retrievals from Linode Object Storage",
		},
	)
	linodeAssetError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "linode_asset_error_total",
			Help: "Number of failed asset retrieval attempts from Linode Object Storage",
		},
	)
)

func init() {
	prometheus.MustRegister(headersHit, memHit, reset, linodeAssetHit, linodeAssetError)
}

type metricsCache struct{}

func (c *metricsCache) HeadersHit()       { headersHit.Inc() }
func (c *metricsCache) HeadersMiss()      { headersMiss.Inc() }
func (c *metricsCache) MemHit()           { memHit.Inc() }
func (c *metricsCache) MemMiss()          { memMiss.Inc() }
func (c *metricsCache) ResetAll()         { reset.Inc() }
func (c *metricsCache) LinodeAssetHit()   { linodeAssetHit.Inc() }
func (c *metricsCache) LinodeAssetError() { linodeAssetError.Inc() }

var Cache metricsCache
