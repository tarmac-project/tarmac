/*
Package logger is a client package for WASM functions running within a Tarmac server.

This package provides user-friendly logging functions that interact with the Tarmac internal logging.
Guest WASM functions running inside Tarmac can import and call this logger.
*/
package logger

import "fmt"

// Logger provides a simple interface for Tarmac Functions to send messages via the standard logger.
type Logger struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// Config provides users with the ability to specify namespaces, function handlers and other key information required to execute the
// function.
type Config struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	Namespace string

	// HostCall is used internally for host callbacks. This is mainly here for testing.
	HostCall func(string, string, string, []byte) ([]byte, error)
}

// New creates a new Logger with the provided configuration.
func New(cfg Config) (*Logger, error) {
	// Set default namespace
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// Verify HostCall is set
	if cfg.HostCall == nil {
		return &Logger{}, fmt.Errorf("HostCall cannot be nil")
	}

	return &Logger{namespace: cfg.Namespace, hostCall: cfg.HostCall}, nil
}

// Trace logs a message at Trace level to the standard logger.
func (l *Logger) Trace(s string) {
	_, _ = l.hostCall(l.namespace, "logger", "trace", []byte(s))
}

// Debug logs a message at Debug level to the standard logger.
func (l *Logger) Debug(s string) {
	_, _ = l.hostCall(l.namespace, "logger", "debug", []byte(s))
}

// Info logs a message at Info level to the standard logger.
func (l *Logger) Info(s string) {
	_, _ = l.hostCall(l.namespace, "logger", "info", []byte(s))
}

// Error logs a message at Error level to the standard logger.
func (l *Logger) Error(s string) {
	_, _ = l.hostCall(l.namespace, "logger", "error", []byte(s))
}
