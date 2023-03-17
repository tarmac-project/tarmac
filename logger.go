package tarmac

// Logger provides a simple interface for Tarmac Functions to send messages via the standard logger.
type Logger struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// newLogger creates a new Logger with the provided configuration.
func newLogger(cfg Config) *Logger {
	return &Logger{namespace: cfg.Namespace, hostCall: cfg.hostCall}
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
