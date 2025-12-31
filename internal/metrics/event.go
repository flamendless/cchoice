package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	clientEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "client",
			Name:      "client_event",
			Help:      "Client event",
		},
		[]string{"event", "value"},
	)
)

func init() {
	prometheus.MustRegister(clientEvent)
}

type metricsClientEvent struct{}

func (c *metricsClientEvent) ClientEventHit(event string, value string) {
	clientEvent.WithLabelValues(event, value)
}

var ClientEvent metricsClientEvent
