package logs

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func InitLog() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	newLogger, err := config.Build()
	if err != nil {
		fmt.Println(err)
	}
	defer newLogger.Sync()
	logger = newLogger
}

func Log() *zap.Logger {
	return logger
}

func LogHTTPHandlerError(r *http.Request, err error) {
	Log().Fatal(
		r.URL.String(),
		zap.String("method", r.Method),
		zap.Error(err),
	)
}
