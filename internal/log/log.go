package log

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

type KurogoLogger struct {
	w     io.Writer
	debug bool
}

var (
	Green = color.New(color.FgGreen, color.Bold)
	Red   = color.New(color.FgRed, color.Bold)
)

// NewKurogoLogger creates a new KurogoLogger.
func NewKurogoLogger(w io.Writer, debug bool) *KurogoLogger {
	l := &KurogoLogger{w: w, debug: debug}
	return l
}

// Printf print log with format.
func (l *KurogoLogger) Printf(c *color.Color, format string, a ...interface{}) {
	var log string
	if c == nil {
		log = fmt.Sprintf(format, a...)
	} else {
		log = c.Sprintf(format, a...)
	}

	fmt.Fprint(l.w, log)
}

// DebugPrintf print log with format when debug is enabled.
func (l *KurogoLogger) DebugPrintf(color *color.Color, format string, a ...interface{}) {
	if l.debug {
		l.Printf(color, format, a...)
	}
}

// EnableDebugLog enable debug log.
func (l *KurogoLogger) EnableDebugLog() {
	l.debug = true
}
