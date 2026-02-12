package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	promoProductImpressions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "promo",
			Name:      "product_impressions_total",
			Help:      "Random product shown",
		},
		[]string{"product_id", "product_name"},
	)
)

func init() {
	prometheus.MustRegister(promoProductImpressions)
}

type metricsPromo struct{}

func (p *metricsPromo) ProductImpressionHit(productID string, productName string) {
	promoProductImpressions.WithLabelValues(productID, productName).Inc()
}

var Promo metricsPromo
