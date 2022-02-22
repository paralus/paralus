package audit

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func getAuditLogger(opts *options) *zap.Logger {

	path := fmt.Sprintf("%s/audit.log", opts.basePath)

	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    opts.maxSizeMB, // megabytes
			MaxBackups: opts.maxBackups,
			MaxAge:     opts.maxAgeDays, // days
		}),
		zap.InfoLevel,
	))

	return logger
}
