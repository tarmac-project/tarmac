/*
Package sdk is a client package for WASM functions running within a Tarmac server.

This package provides user-friendly functions that wrap the Web Assembly Procedure Call (waPC) based functions of
Tarmac. Guest WASM functions running inside Tarmac can use this library to call back the Tarmac host and perform
host-level actions such as storing data within the database, logging specific data, or looking up configurations.

	import "github.com/madflojo/tarmac/pkg/sdk"
	
	var tarmac *Tarmac
	
	func main() {
		tarmac, err := sdk.New(sdk.Config{Namespace: "my-service", Handler: Handler})
		if err != nil {
			// do something
		}
	}
	
	func Handler(payload []byte) ([]byte, error) {
		tarmac.Logger.Info("This is a log message")
		return []byte("Hello World"), nil
	}

*/
package sdk

import (
	"fmt"
	"github.com/madflojo/tarmac/pkg/sdk/logger"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

// Tarmac provides an interface to users which wraps and simplifies the interfaces for WASM Functions execute by Tarmac. This interface
// provides access to Loggers, Metrics, KVStores, and SQL Databases.
type Tarmac struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	namespace string

	// Handler registers the user function to execute as part of the Tarmac Function.
	handler func([]byte) ([]byte, error)

	Logger *logger.Logger
}

// Config provides users with the ability to specify namespaces, function handlers and other key information required to execute the
// function.
type Config struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	Namespace string

	// Handler registers the user function to execute as part of the Tarmac Function.
	Handler func([]byte) ([]byte, error)

	// hostCall is used internally for host callbacks. This is mainly here for testing.
	hostCall func(string, string, string, []byte) ([]byte, error)
}

// New creates a new Tarmac instance with the specified configuration.
func New(cfg Config) (*Tarmac, error) {
	t := &Tarmac{namespace: "default"}
	if cfg.Namespace != "" {
		t.namespace = cfg.Namespace
	}

	// Validate Handler is not empty
	if cfg.Handler == nil {
		return t, fmt.Errorf("function handler cannot be nil")
	}

	// Register provided handler
	wapc.RegisterFunctions(wapc.Functions{
		"handler": cfg.Handler,
	})

	// Set hostCall function for internal callbacks
	cfg.hostCall = wapc.HostCall

	var err error

	// Initialize a Logger instance
	t.Logger, err = logger.NewLogger(logger.Config{Namespace: cfg.Namespace, HostCall: cfg.hostCall})
	if err != nil {
		return t, fmt.Errorf("error while initializing logger - %s", err)
	}

	return t, nil
}
