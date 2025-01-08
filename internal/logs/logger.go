package logs

import (
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	InitLog()
}

func InitLog() {
	var config zap.Config
	env := os.Getenv("APP_ENV")
	if env == "prod" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	newLogger, err := config.Build()
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err := newLogger.Sync()
		if err != nil {
			fmt.Println(err)
		}
	}()
	logger = newLogger
}

func Log() *zap.Logger {
	return logger
}

func LogHTTPHandler(logger *zap.Logger, r *http.Request, err error) {
	logger.Warn(
		r.URL.String(),
		zap.String("method", r.Method),
		zap.Error(err),
	)
}
