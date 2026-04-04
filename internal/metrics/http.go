package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "http",
			Name:      "errors_total",
			Help:      "Total HTTP error responses",
		},
		[]string{"method", "path", "status"},
	)
	httpRoutesSkippedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "http",
			Name:      "routes_skipped_total",
			Help:      "Total HTTP routes skipped",
		},
		[]string{"method", "path"},
	)
	httpRateLimitedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "http",
			Name:      "rate_limited_total",
			Help:      "Total requests blocked by rate limiting",
		},
		[]string{"path", "ip"},
	)
	httpRateLimitActiveVisitors = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "cchoice",
			Subsystem: "http",
			Name:      "rate_limit_active_visitors",
			Help:      "Current number of active rate limit visitors",
		},
	)
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpErrorsTotal,
		httpRoutesSkippedTotal,
		httpRateLimitedTotal,
		httpRateLimitActiveVisitors,
	)
}

type metricsHTTP struct{}

func (h *metricsHTTP) RequestsHit(params ...string) {
	httpRequestsTotal.WithLabelValues(params...).Inc()
}
func (h *metricsHTTP) ErrorsHit(params ...string) { httpErrorsTotal.WithLabelValues(params...).Inc() }
func (h *metricsHTTP) RoutesSkippedHit(params ...string) {
	httpRoutesSkippedTotal.WithLabelValues(params...).Inc()
}
func (h *metricsHTTP) RateLimitedHit(params ...string) {
	httpRateLimitedTotal.WithLabelValues(params...).Inc()
}
func (h *metricsHTTP) SetRateLimitActiveVisitors(v float64) { httpRateLimitActiveVisitors.Set(v) }

var HTTP metricsHTTP
