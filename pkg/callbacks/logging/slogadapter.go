package logging

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	// LevelTrace is the trace level
	LevelTrace = slog.LevelDebug - 4
)

// SlogAdapter implements the logging.Log interface for slog.Logger
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new adapter that wraps a slog.Logger
func NewSlogAdapter(logger *slog.Logger) Log {
	return &SlogAdapter{logger: logger}
}

// Info logs at info level
func (l *SlogAdapter) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

// Warn logs at warn level
func (l *SlogAdapter) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

// Error logs at error level
func (l *SlogAdapter) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

// Debug logs at debug level
func (l *SlogAdapter) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

// Trace logs at debug level with a trace marker
func (l *SlogAdapter) Trace(args ...interface{}) {
	l.logger.Log(context.Background(), LevelTrace, fmt.Sprint(args...))
}
