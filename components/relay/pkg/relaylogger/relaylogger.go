package relaylogger

import (
	"bytes"
	"fmt"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"go.uber.org/zap"
)

var (
	logtmp   = log.GetLogger()
	plain    = logtmp.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(2))
	rlog     = plain.Sugar()
	runLevel = 0
)

// RelayLog is a sample logr.Logger that logs to stderr.
// It's terribly inefficient, and is *only* a basic example.
type RelayLog struct {
	level     int
	name      string
	keyValues map[string]interface{}
}

func (l *RelayLog) relaylog(level int, msg string, kvs ...interface{}) {
	var buf bytes.Buffer

	if level > runLevel {
		return
	}

	fmt.Fprintf(&buf, "%s::%s ", l.name, msg)
	for k, v := range l.keyValues {
		fmt.Fprintf(&buf, "%s: %+v  ", k, v)
	}

	if len(kvs)%2 == 0 {
		for i := 0; i < len(kvs); i += 2 {
			fmt.Fprintf(&buf, "%s: %+v  ", kvs[i], kvs[i+1])

		}
	} else {
		for i := 0; i < len(kvs); i++ {
			fmt.Fprintf(&buf, "%+v ", kvs[i])
		}
	}

	fmt.Fprintf(&buf, "\n")
	switch level {
	case 0:
		rlog.Panicw(buf.String())
	case 1:
		rlog.Errorw(buf.String())
	case 2:
		rlog.Warnw(buf.String())
	case 3:
		rlog.Infow(buf.String())
	case 4:
		rlog.Debugw(buf.String())
	}

}

//Enabled log enabled
func (_ *RelayLog) Enabled() bool {
	return true
}

//Critical log critical
func (l *RelayLog) Critical(msg string, kvs ...interface{}) {
	l.relaylog(0, msg, kvs...)
}

//Error log error level
func (l *RelayLog) Error(err error, msg string, kvs ...interface{}) {
	kvs = append(kvs, "error", err)
	l.relaylog(1, msg, kvs...)
}

//Warn log warning level
func (l *RelayLog) Warn(msg string, kvs ...interface{}) {
	l.relaylog(2, msg, kvs...)
}

//Info log info level
func (l *RelayLog) Info(msg string, kvs ...interface{}) {
	l.relaylog(3, msg, kvs...)
}

//Debug log debug level
func (l *RelayLog) Debug(msg string, kvs ...interface{}) {
	l.relaylog(4, msg, kvs...)
}

//V ..
func (l *RelayLog) V(_ int) *RelayLog {
	return l
}

//WithName set log name
func (l *RelayLog) WithName(name string) *RelayLog {
	var objName string
	if l.name == "" {
		objName = name
	} else {
		objName = l.name + "." + name
	}

	return &RelayLog{
		level:     l.level,
		name:      objName,
		keyValues: l.keyValues,
	}
}

//WithValues log  key values
func (l *RelayLog) WithValues(kvs ...interface{}) *RelayLog {
	newMap := make(map[string]interface{}, len(l.keyValues)+len(kvs)/2)
	for k, v := range l.keyValues {
		newMap[k] = v
	}
	for i := 0; i < len(kvs); i += 2 {
		newMap[kvs[i].(string)] = kvs[i+1]
	}
	return &RelayLog{
		level:     l.level,
		name:      l.name,
		keyValues: newMap,
	}
}

//NewLogger create new logger
func NewLogger(level int) *RelayLog {
	if runLevel < level {
		runLevel = level
	}
	return &RelayLog{level: level}
}

//SetRunTimeLogLevel set a run time log level
func SetRunTimeLogLevel(level int) {
	fmt.Println("log level changed from ", runLevel, " to ", level)
	runLevel = level
}
