package internal

import (
	cchoice_db "cchoice/cchoice_db"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

type AppFlags struct {
	Template               string
	Filepath               string
	Sheet                  string
	Strict                 bool
	Limit                  int
	PrintProcessedProducts bool
	VerifyPrices           bool
	DBPath                 string
	UseDB                  bool
}

type Metrics struct {
	Time []zap.Field
}

type AppContext struct {
	DB      *sql.DB
	Queries *cchoice_db.Queries
	Metrics *Metrics
}

func (metrics *Metrics) Add(key string, dur time.Duration) {
	metrics.Time = append(metrics.Time, zap.Duration(key, dur))
}

func (metrics *Metrics) LogTime(log *zap.Logger) {
	log.Info("Time Metrics", metrics.Time...)
}
