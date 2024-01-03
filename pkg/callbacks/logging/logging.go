/*
Package logging is part of the Tarmac suite of Host Callback packages. This package provides users with the ability
to provide WASM functions with a host callback interface that provides logging capabilities.

	import (
		"github.com/tarmac-project/tarmac/pkg/callbacks/logging"
	)

	func main() {
		// Create instance of logger to register for callback execution
		logger := logging.New(logging.Config{
			Log: MyCustomLogger,
		})

		// Create Callback router and register logger
		router := callbacks.New()
		router.RegisterCallback("logging", "Info", logger.Info)
	}
*/
package logging

// Logger provides a Host Callback interface to interact with the underlying logging system. WASM functions can call
// this callback interface which will coordinate and execute the host-supplied logger instance.
type Logger struct {
	log Log
}

// Log is an interface that can be provided and satisfied by the instantiator of the Logger struct. This interface
// allows users to specify their logging functionality or incorporate an existing logger into the Tarmac Host
// Callback functions.
type Log interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})
}

// Config is provided to users to configure the Host Callback. All Tarmac Callbacks follow the same configuration
// format; each Config struct gives the specific Host Callback unique functionality.
type Config struct {
	// Log is an interface that can be provided by the instantiator. This interface allows users to specify their logging
	// functionality or incorporate an existing logger into the Tarmac Host Callback function.
	Log Log
}

// New will create and return a new logger instance that users can register as a Tarmac Host Callback function. Users
// can provide a custom logger using the Configuration options supplied. By default, if there is no custom logger
// provided, the NoOpLog will be used.
func New(cfg Config) (*Logger, error) {
	// Create default logger
	l := &Logger{
		log: NoOpLog{},
	}

	// If provided, override with user logger
	if cfg.Log != nil {
		l.log = cfg.Log
	}

	return l, nil
}

// Info will take the incoming byte slice data and call the internal logger converting the data to a string.
func (l *Logger) Info(b []byte) ([]byte, error) {
	l.log.Info(string(b))
	return []byte(""), nil
}

// Error will take the incoming byte slice data and call the internal logger converting the data to a string.
func (l *Logger) Error(b []byte) ([]byte, error) {
	l.log.Error(string(b))
	return []byte(""), nil
}

// Debug will take the incoming byte slice data and call the internal logger converting the data to a string.
func (l *Logger) Debug(b []byte) ([]byte, error) {
	l.log.Debug(string(b))
	return []byte(""), nil
}

// Trace will take the incoming byte slice data and call the internal logger converting the data to a string.
func (l *Logger) Trace(b []byte) ([]byte, error) {
	l.log.Trace(string(b))
	return []byte(""), nil
}

// Warn will take the incoming byte slice data and call the internal logger converting the data to a string.
func (l *Logger) Warn(b []byte) ([]byte, error) {
	l.log.Warn(string(b))
	return []byte(""), nil
}

// NoOpLog is a convience logger which is used as a default when users do not provide their own interface.
type NoOpLog struct{}

// Info does nothing
func (log NoOpLog) Info(...interface{}) {}

// Warn does nothing
func (log NoOpLog) Warn(...interface{}) {}

// Error does nothing
func (log NoOpLog) Error(...interface{}) {}

// Debug does nothing
func (log NoOpLog) Debug(...interface{}) {}

// Trace does nothing
func (log NoOpLog) Trace(...interface{}) {}
