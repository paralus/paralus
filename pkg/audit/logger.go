package audit

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// options holds audit options
type AuditOptions struct {
	LogPath    string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

func GetAuditLogger(opts *AuditOptions) *zap.Logger {
	encoder := zapcore.EncoderConfig{
		TimeKey:    "timestamp",
		EncodeTime: zapcore.RFC3339NanoTimeEncoder,
	}
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   opts.LogPath,
			MaxSize:    opts.MaxSizeMB, // megabytes
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAgeDays, // days
		}),
		zap.InfoLevel,
	))

	return logger
}
