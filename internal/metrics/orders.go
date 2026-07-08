package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ordersCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "orders",
			Name:      "created_total",
			Help:      "Total orders created from checkout",
		},
		[]string{"payment_method"},
	)
	ordersPaidTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "orders",
			Name:      "paid_total",
			Help:      "Total orders paid successfully",
		},
	)
	ordersCancelledTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "orders",
			Name:      "cancelled_total",
			Help:      "Total orders cancelled after payment redirect",
		},
	)
)

func init() {
	prometheus.MustRegister(ordersCreatedTotal, ordersPaidTotal, ordersCancelledTotal)
}

type metricsOrders struct{}

func (o *metricsOrders) Created(paymentMethod string) {
	ordersCreatedTotal.WithLabelValues(paymentMethod).Inc()
}

func (o *metricsOrders) Paid() { ordersPaidTotal.Inc() }

func (o *metricsOrders) Cancelled() { ordersCancelledTotal.Inc() }

var Orders metricsOrders
