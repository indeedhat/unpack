package generator

import (
	"fmt"
	"os"
	"strings"
)

type logger interface {
	WithArgs(msg string, args ...any)
	Log(msg string)
	Logf(msg string, args ...any)
}

type consoleLogger struct{}

// Log implements [logger].
func (consoleLogger) Log(msg string) {
	fmt.Fprint(os.Stdout, msg+"\n")
}

func (consoleLogger) Logf(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, msg+"\n", args...)
}

// WithArgs implements [logger].
func (consoleLogger) WithArgs(msg string, args ...any) {
	pattern := strings.Repeat("    %s = %v\n", len(args)/2)

	fmt.Fprintf(os.Stdout,
		fmt.Sprintf("%s:\n%s", msg, pattern),
		args...,
	)
}

var _ logger = (*consoleLogger)(nil)

type noopLogger struct{}

// Log implements [logger].
func (noopLogger) Log(string) {}

// Logf implements [logger].
func (noopLogger) Logf(string, ...any) {}

// WithArgs implements [logger].
func (noopLogger) WithArgs(string, ...any) {}

var _ logger = (*noopLogger)(nil)
