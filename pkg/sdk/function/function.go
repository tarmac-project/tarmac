/*
Package function is a client package for WASM functions running within a Tarmac server.

This package provides a user-friendly interface for interacting with other WASM functions. Users can use this package to call other loaded WASM functions. This capability enables users to create and call functions as if they are internal only.
*/
package function

import "fmt"

// Function provides a simple interface for Tarmac Functions to send messages via the standard function.
type Function struct {
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

// New creates a new Function with the provided configuration.
func New(cfg Config) (*Function, error) {
	// Set default namespace
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// Verify HostCall is set
	if cfg.HostCall == nil {
		return &Function{}, fmt.Errorf("HostCall cannot be nil")
	}

	return &Function{namespace: cfg.Namespace, hostCall: cfg.HostCall}, nil
}

// Call will call other functions registered via Tarmac routes.
func (f *Function) Call(name string, input []byte) ([]byte, error) {
	return f.hostCall(f.namespace, "function", name, input)
}
