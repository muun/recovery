package utils

import (
	"fmt"
	"os"
	"strings"
)

// DebugMode is true when the `DEBUG` environment variable is set to "true".
var DebugMode bool = os.Getenv("DEBUG") == "true"

// Logger provides logging methods that only print when `DebugMode` is true.
// This allows callers to log detailed information without displaying it to users during normal
// execution.
type Logger struct {
	tag string
}

// NewLogger returns an initialized Logger instance.
func NewLogger(tag string) *Logger {
	return &Logger{tag}
}

// SetTag updates the tag of this Logger.
func (l *Logger) SetTag(newTag string) {
	l.tag = newTag
}

// Printf works like fmt.Printf, but only prints when `DebugMode` is true.
func (l *Logger) Printf(format string, v ...interface{}) {
	if !DebugMode {
		return
	}

	message := strings.TrimSpace(fmt.Sprintf(format, v...))

	fmt.Printf("%s %s\n", l.getPrefix(), message)
}

// Errorf works like fmt.Errorf, but prints the error to the console if `DebugMode` is true.
func (l *Logger) Errorf(format string, v ...interface{}) error {
	err := fmt.Errorf(format, v...)

	if DebugMode {
		fmt.Printf("%s %v\n", l.getPrefix(), err)
	}

	return err
}

func (l *Logger) getPrefix() string {
	return fmt.Sprintf("[%s]", l.tag)
}
