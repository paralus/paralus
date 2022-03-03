package log

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	log := GetLogger()

	lc := make(chan string)

	go log.ChangeLevel(lc)

	log.Infow("test ingo", "key", "value")
	log.Debugw("test debug", "key", "value")
	log.Warn("changing to debug")
	lc <- "debug"
	time.Sleep(time.Second)
	log.Infow("test ingo", "key", "value")
	log.Debugw("test debug", "key", "value")

}
