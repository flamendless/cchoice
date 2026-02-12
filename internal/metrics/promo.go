package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	promoProductImpressions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "promo",
			Name:      "product_impressions_total",
			Help:      "Total number of times a random promo product was shown",
		},
		[]string{"product_id", "datetime"},
	)
)

func init() {
	prometheus.MustRegister(promoProductImpressions)
}

type metricsPromo struct{}

func (p *metricsPromo) ProductImpressionHit(productID string) {
	dt := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	promoProductImpressions.WithLabelValues(productID, dt).Inc()
}

var Promo metricsPromo
