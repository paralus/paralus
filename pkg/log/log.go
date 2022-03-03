package log

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level is the level of logging for logger
type level string

// LogLevel constants
const (
	Debug level = "debug"
	Info  level = "info"
	Warn  level = "warn"
	Error level = "error"
)

func newConfig(l level) zap.Config {
	var zl zapcore.Level
	switch l {
	case Debug:
		zl = zap.DebugLevel
	case Warn:
		zl = zap.WarnLevel
	case Info:
		zl = zap.InfoLevel
	case Error:
		zl = zap.ErrorLevel
	default:
		zl = zap.InfoLevel
	}

	return zap.Config{
		Level:    zap.NewAtomicLevelAt(zl),
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			CallerKey:      "caller",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
}

// Logger is a wrapper around zap sugared logger
type Logger struct {
	*zap.SugaredLogger
}

// ChangeLevel changes the log level of logger
func (lg *Logger) ChangeLevel(lc <-chan string) {
	for {
		select {
		case lvl := <-lc:
			l := level(lvl)
			cfg := newConfig(l)
			zl, err := cfg.Build()
			if err != nil {
				panic(err)
			}

			lg.SugaredLogger = zl.Sugar()
		}
	}
}

var zl *Logger
var lo sync.Once

// GetLogger return instance of logger
func GetLogger() *Logger {
	if zl == nil {
		lo.Do(func() {
			cfg := newConfig(Info)
			z, err := cfg.Build()
			if err != nil {
				panic(err)
			}
			zl = &Logger{
				SugaredLogger: z.Sugar(),
			}

		})
	}
	return zl
}
