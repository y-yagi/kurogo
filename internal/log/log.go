package log

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

type PorterLogger struct {
	w     io.Writer
	debug bool
}

var (
	Green = color.New(color.FgGreen, color.Bold)
	Red   = color.New(color.FgRed, color.Bold)
)

// NewPorterLogger creates a new PorterLogger.
func NewPorterLogger(w io.Writer, debug bool) *PorterLogger {
	l := &PorterLogger{w: w, debug: debug}
	return l
}

// Printf print log with format.
func (l *PorterLogger) Printf(c *color.Color, format string, a ...interface{}) {
	var log string
	if c == nil {
		log = fmt.Sprintf(format, a...)
	} else {
		log = c.Sprintf(format, a...)
	}

	fmt.Fprint(l.w, log)
}

// DebugPrintf print log with format when debug is enabled.
func (l *PorterLogger) DebugPrintf(color *color.Color, format string, a ...interface{}) {
	if l.debug {
		l.Printf(color, format, a)
	}
}
