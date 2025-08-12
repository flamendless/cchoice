package models

import (
	"time"

	"go.uber.org/zap"
)

type Metrics struct {
	Time []zap.Field
}

func (metrics *Metrics) Add(key string, dur time.Duration) {
	metrics.Time = append(metrics.Time, zap.Duration(key, dur))
}

func (metrics *Metrics) LogTime(log *zap.Logger) {
	log.Info("Time Metrics", metrics.Time...)
}
