package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// DebugMode is true when the `DEBUG` environment variable is set to "true".
var DebugMode bool = os.Getenv("DEBUG") == "true"
var outputStream io.Writer = io.Discard

// SetOutputStream set a writer to record all logs (except Tracef).
func SetOutputStream(s io.Writer) {
	outputStream = s
}

// Logger provides logging methods. Logs are written to the stream set by
// SetOutputStream. If `DebugMode` is true, we also print to stdout/stderr.
// This allows callers to log detailed information without displaying it to
// users during normal execution.
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

// Tracef works like fmt.Printf, but only prints when `DebugMode` is true. These logs
// are *not* recorded to the output stream.
func (l *Logger) Tracef(format string, v ...interface{}) {
	if !DebugMode {
		return
	}

	message := strings.TrimSpace(fmt.Sprintf(format, v...))

	fmt.Printf("%s %s %s\n", time.Now().Format(time.RFC3339Nano), l.getPrefix(), message)
}

// Printf works like fmt.Printf, but only prints when `DebugMode` is true. These logs
// are recorded to the output stream, so they should not include sensitive information.
func (l *Logger) Printf(format string, v ...interface{}) {
	message := strings.TrimSpace(fmt.Sprintf(format, v...))

	log := fmt.Sprintf("%s %s %s\n", time.Now().Format(time.RFC3339Nano), l.getPrefix(), message)
	_, _ = outputStream.Write([]byte(log))
	if DebugMode {
		print(log)
	}
}

// Errorf works like fmt.Errorf, but prints the error to the console if `DebugMode` is true.
// These logs are recorded to the output stream, so they should not include sensitive information.
func (l *Logger) Errorf(format string, v ...interface{}) error {
	err := fmt.Errorf(format, v...)

	log := fmt.Sprintf("ERROR: %s %s %v\n", time.Now().Format(time.RFC3339Nano), l.getPrefix(), err)
	_, _ = outputStream.Write([]byte(log))
	if DebugMode {
		print(log)
	}

	return err
}

func (l *Logger) getPrefix() string {
	return fmt.Sprintf("[%s]", l.tag)
}
