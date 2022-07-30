package utils

import (
	"testing"

	"github.com/shake-on-it/app-tmpl/backend/common"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(t *testing.T) common.Logger {
	logger, err := common.NewLogger("test", common.LoggerOptionsProd)
	if err != nil {
		t.Logf("failed to create test logger: %s", err)
		t.Fail()
		return nil
	}

	return logger.
		Desugar().
		WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core { return zapcore.NewNopCore() })).
		WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, zapcore.NewCore(
				zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
					NameKey:        "logger",
					LevelKey:       "level",
					TimeKey:        "time",
					CallerKey:      "caller",
					MessageKey:     "msg",
					StacktraceKey:  "stacktrace",
					EncodeLevel:    zapcore.CapitalColorLevelEncoder,
					EncodeTime:     zapcore.ISO8601TimeEncoder,
					EncodeDuration: zapcore.StringDurationEncoder,
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}),
				zapcore.AddSync(logWriter{t.Logf}),
				zapcore.DebugLevel,
			))
		})).
		Sugar()
}

type logWriter struct {
	logf func(msg string, args ...interface{})
}

func (l logWriter) Logf(msg string, args ...interface{}) {
	l.logf(msg, args...)
}

func (l logWriter) Write(p []byte) (n int, err error) {
	l.logf("%s", string(p))
	return len(p), nil
}
