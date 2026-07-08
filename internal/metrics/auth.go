package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	AuthUserTypeAdmin    = "admin"
	AuthUserTypeCustomer = "customer"
	AuthResultSuccess    = "success"
	AuthResultFailure    = "failure"
)

var (
	authLoginAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "auth",
			Name:      "login_attempts_total",
			Help:      "Total login attempts",
		},
		[]string{"user_type", "result"},
	)
	authRegistrationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cchoice",
			Subsystem: "auth",
			Name:      "registrations_total",
			Help:      "Total customer registrations",
		},
		[]string{"result"},
	)
)

func init() {
	prometheus.MustRegister(authLoginAttemptsTotal, authRegistrationsTotal)
}

type metricsAuth struct{}

func (a *metricsAuth) LoginAttempt(userType, result string) {
	authLoginAttemptsTotal.WithLabelValues(userType, result).Inc()
}

func (a *metricsAuth) Registration(result string) {
	authRegistrationsTotal.WithLabelValues(result).Inc()
}

var Auth metricsAuth
