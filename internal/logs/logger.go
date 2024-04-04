package logs

import (
	"fmt"

	"go.uber.org/zap"
)

var logger *zap.Logger

func InitLog() {
	newLogger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println(err)
	}
	defer newLogger.Sync()
	logger = newLogger
}

func Log() *zap.Logger {
	return logger
}
