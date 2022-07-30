package common

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LoggerFieldDuration   = "duration"
	LoggerFieldHTTPMethod = "http_method"
	LoggerFieldIPAddress  = "ip_address"
	LoggerFieldPath       = "path"
	LoggerFieldProto      = "proto"
	LoggerFieldRequestID  = "request_id"
	LoggerFieldHTTPStatus = "http_status"
)

// Logger is a logger
type Logger = *zap.SugaredLogger

func NewLogger(name string, opts LoggerOptions) (Logger, error) {
	if opts.Encoding == "" {
		opts.Encoding = LoggerEncodingConsole
	}

	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.DebugLevel)

	config := zap.Config{
		Level:    level,
		Encoding: opts.Encoding,
		EncoderConfig: zapcore.EncoderConfig{
			NameKey:        "logger",
			LevelKey:       "level",
			TimeKey:        "time",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    opts.Level,
			EncodeTime:     opts.Time,
			EncodeDuration: opts.Duration,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		DisableStacktrace: true,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger.Sugar().Named(name), nil
}

// set of supported logger encodings
const (
	LoggerEncodingConsole = "console"
	LoggerEncodingJSON    = "json"
)

// LoggerOptions is the logger options
type LoggerOptions struct {
	Duration zapcore.DurationEncoder
	Encoding string
	Level    zapcore.LevelEncoder
	Time     zapcore.TimeEncoder
}

// set of logger pre-configurations
var (
	LoggerOptionsDev = LoggerOptions{
		Encoding: LoggerEncodingConsole,
		Level:    zapcore.CapitalColorLevelEncoder,
		Time:     zapcore.ISO8601TimeEncoder,
		Duration: zapcore.StringDurationEncoder,
	}

	LoggerOptionsProd = LoggerOptions{
		Encoding: LoggerEncodingJSON,
		Level:    zapcore.LowercaseLevelEncoder,
		Time:     zapcore.EpochTimeEncoder,
		Duration: zapcore.SecondsDurationEncoder,
	}
)
