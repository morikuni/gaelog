package gaelog

import (
	"context"

	"google.golang.org/appengine/log"
)

// CustomLogger is a logger for the standard environment.
type StandardLogger struct{}

// Criticalf implements Logger.
func (l StandardLogger) Criticalf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Critical, format, args...)
}

// Errorf implements Logger.
func (l StandardLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Error, format, args...)
}

// Warningf implements Logger.
func (l *StandardLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Warning, format, args...)
}

// Infof implements Logger.
func (l *StandardLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Info, format, args...)
}

// Debugf implements Logger.
func (l *StandardLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Printf(ctx, Debug, format, args...)
}

// Printf implements Logger.
func (l StandardLogger) Printf(ctx context.Context, level LogLevel, format string, args ...interface{}) {
	switch level {
	case Critical:
		log.Criticalf(ctx, format, args...)
	case Error:
		log.Errorf(ctx, format, args...)
	case Warning:
		log.Warningf(ctx, format, args...)
	case Info:
		log.Infof(ctx, format, args...)
	case Debug:
		log.Debugf(ctx, format, args...)
	}
}
