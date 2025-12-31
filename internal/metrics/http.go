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
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpErrorsTotal,
		httpRoutesSkippedTotal,
	)
}

type metricsHTTP struct{}

func (h *metricsHTTP) RequestsHit(params ...string) { httpRequestsTotal.WithLabelValues(params...) }
func (h *metricsHTTP) ErrorsHit(params ...string) { httpErrorsTotal.WithLabelValues(params...) }
func (h *metricsHTTP) RoutesSkippedHit(params ...string) { httpRoutesSkippedTotal.WithLabelValues(params...) }

var HTTP metricsHTTP
