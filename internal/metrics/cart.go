package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	CheckoutResultSuccess = "success"
	CheckoutResultFailure = "failure"
)

var (
	cartCheckoutAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "cart",
			Name:      "checkout_attempts_total",
			Help:      "Total cart checkout attempts",
		},
		[]string{"result"},
	)
)

func init() {
	prometheus.MustRegister(cartCheckoutAttemptsTotal)
}

type metricsCart struct{}

func (c *metricsCart) CheckoutAttempt(result string) {
	cartCheckoutAttemptsTotal.WithLabelValues(result).Inc()
}

var Cart metricsCart
