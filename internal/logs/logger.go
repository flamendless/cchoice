package logs

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/errs"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"syscall"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var loggerOnce sync.Once
var logger *zap.Logger

func InitLog() {
	var config zap.Config

	cfg := conf.Conf()
	switch {
	case cfg.IsLocal():
		config = zap.NewDevelopmentConfig()
		configLevel := conf.Conf().LogMinLevel
		config.Level.SetLevel(zapcore.Level(configLevel - 1))
	case cfg.IsProd():
		config = zap.NewProductionConfig()
	default:
		panic(fmt.Errorf("%w. APP_ENV", errs.ErrEnvVarRequired))
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

func JSONResponse(id string, resp *http.Response) {
	if resp == nil {
		logger.Error("Passed a nil *https.Response")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Log().Error("Failed to read response body", zap.Error(err))
		return
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var prettyBuf bytes.Buffer
	if err := json.Indent(&prettyBuf, body, "", "  "); err != nil {
		Log().Error("Failed to pretty-print JSON", zap.Error(err))
		return
	}

	Log().Sugar().Info("Pretty JSON", zap.String("body", prettyBuf.String()))
}

func CacheMiss(key []byte) {
	Log().Debug(constants.CacheMiss, zap.ByteString("key", key))
}

func CacheHit(key []byte, length int) {
	Log().Debug(
		constants.CacheHit,
		zap.ByteString("key", key),
		zap.Int("len", length),
	)
}

func CacheStore(key []byte, buf *bytes.Buffer) {
	Log().Debug(
		constants.CacheStore,
		zap.ByteString("key", key),
		zap.Int("len", buf.Len()),
	)
}

func GobError(key []byte, err error) {
	Log().Debug(
		constants.Gob,
		zap.ByteString("key", key),
		zap.Error(err),
	)
}

func SF(key []byte, shared bool) {
	Log().Debug(
		constants.Singleflight,
		zap.ByteString("key", key),
		zap.Bool("shared", shared),
	)
}
