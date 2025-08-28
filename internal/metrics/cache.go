package metrics

import (
	"sync/atomic"
)

var Cache metricsCache

type metricsCache struct {
	headersHit  atomic.Int64
	headersMiss atomic.Int64
	clientHit   atomic.Int64
	clientMiss  atomic.Int64
}

func (c *metricsCache) HitHeaders() {
	c.headersHit.Add(1)
}

func (c *metricsCache) MissHeaders() {
	c.headersMiss.Add(1)
}

func (c *metricsCache) HitClient() {
	c.clientHit.Add(1)
}

func (c *metricsCache) MissClient() {
	c.clientMiss.Add(1)
}

func (c *metricsCache) ValueHitHeaders() int64 {
	return c.headersHit.Load()
}

func (c *metricsCache) ValueHitClient() int64 {
	return c.clientHit.Load()
}

func (c *metricsCache) ValueMissHeaders() int64 {
	return c.headersMiss.Load()
}

func (c *metricsCache) ValueMissClient() int64 {
	return c.clientMiss.Load()
}

func (c *metricsCache) ResetAll() {
	c.headersHit.Store(0)
	c.headersMiss.Store(0)
	c.clientHit.Store(0)
	c.clientMiss.Store(0)
}
