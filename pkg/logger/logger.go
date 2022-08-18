// Package logger implements context-aware logging.
//
// It adds tracing into HTTP communication and allows us to track concurrent
// log records by request ID. It may be handy since the log is non-linear because
// goroutines add logs asynchronously.
package logger

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	CtxLoggerKey string

	Logger struct {
		*zap.Logger
	}
)

const LoggerKey CtxLoggerKey = "logger"

var fallbackLogger *zap.SugaredLogger

// Logging function (a Zap wrapper) which considers context.
// Usage example: `Log(r.Context()).Errorw("...")` etc. See the Zap docs.
func Log(ctx context.Context) *zap.SugaredLogger {
	if ctx == nil {
		return fallbackLogger
	}
	zap, ok := ctx.Value(LoggerKey).(*zap.SugaredLogger)
	if !ok || zap == nil {
		return fallbackLogger
	}
	return zap
}

func Run(level string) *Logger {
	var zapLogger *zap.Logger

	var minLevel zapcore.Level
	switch level {
	case "debug":
		minLevel = zap.DebugLevel
	case "info":
		minLevel = zap.InfoLevel
	case "warn":
		minLevel = zap.WarnLevel
	case "error":
		minLevel = zapcore.ErrorLevel
	case "dpanic":
		minLevel = zap.DPanicLevel
	case "panic":
		minLevel = zap.PanicLevel
	case "fatal":
		minLevel = zap.FatalLevel
	default:
		minLevel = zap.ErrorLevel
	}

	zapLogger, err := zap.NewDevelopment()
	zapLogger.With(
		zap.String("logger", "ctxLog"),
	).WithOptions(
		zap.IncreaseLevel(minLevel),
	).Sugar()

	if err != nil {
		log.Fatal("logger: can't init Zap logger")
	}
	defer zapLogger.Sync()

	logger := &Logger{zapLogger}

	fallbackLogger = zapLogger.With(
		zap.String("logger", "fallbackLogger"),
	).WithOptions(
		zap.IncreaseLevel(minLevel),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.DebugLevel),
	).Sugar()

	return logger
}
