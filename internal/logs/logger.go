package logs

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var loggerOnce sync.Once
var logger *zap.Logger

func InitLog() {
	var config zap.Config
	env := os.Getenv("APP_ENV")
	switch env {
	case "prod":
		config = zap.NewProductionConfig()
	case "local":
		config = zap.NewDevelopmentConfig()
	default:
		panic("Invalid app env")
	}

	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	newLogger, err := config.Build()
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := newLogger.Sync(); !errors.Is(err, syscall.EINVAL) {
			fmt.Println(err)
		}
	}()
	logger = newLogger
}

func Log() *zap.Logger {
	loggerOnce.Do(func() {
		InitLog()
	})
	return logger
}

func LogHTTPHandler(logger *zap.Logger, r *http.Request, err error) {
	logger.Warn(
		r.URL.String(),
		zap.String("method", r.Method),
		zap.Error(err),
	)
}
