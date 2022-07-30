package common

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

const (
	panicTimestampFormat = "%d-%02d-%02dT%02d:%02d:%02d.%03d+0000\n"
)

func WritePanic(err interface{}, logger Logger) {
	now := time.Now()
	fmt.Fprintf(os.Stderr,
		panicTimestampFormat,
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1e6,
	)

	errMsg := fmt.Sprintf("panic occurred: %v", err)
	fmt.Fprintf(os.Stderr, errMsg)

	debug.PrintStack()

	if logger != nil {
		logger.Error(fmt.Sprintf("%s\n%s", errMsg, debug.Stack()))
	}
}
