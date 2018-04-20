package gaelog

import (
	"context"
)

// LogLevel is a enum of the leg level.
type LogLevel string

const (
	// Critical is critical log level.
	Critical LogLevel = "CRITICAL"
	// Error is error log level.
	Error LogLevel = "ERROR"
	// Warning is error log level.
	Warning LogLevel = "WARNING"
	// Info is info log level.
	Info LogLevel = "INFO"
	// Debug is debug log level.
	Debug LogLevel = "DEBUG"
)

// Logger represents a logger on GAE.
type Logger interface {
	Criticalf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Warningf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
	Printf(ctx context.Context, level LogLevel, format string, args ...interface{})
}
