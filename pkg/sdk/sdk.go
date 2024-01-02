/*
Package sdk is a client package for WASM functions running within a Tarmac server.

This package provides user-friendly functions that wrap the Web Assembly Procedure Call (waPC) based functions of
Tarmac. Guest WASM functions running inside Tarmac can use this library to call back the Tarmac host and perform
host-level actions such as storing data within the database, logging specific data, or looking up configurations.

	import "github.com/tarmac-project/tarmac/pkg/sdk"

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

	"github.com/tarmac-project/tarmac/pkg/sdk/function"
	"github.com/tarmac-project/tarmac/pkg/sdk/http"
	"github.com/tarmac-project/tarmac/pkg/sdk/kvstore"
	"github.com/tarmac-project/tarmac/pkg/sdk/logger"
	"github.com/tarmac-project/tarmac/pkg/sdk/metrics"
	"github.com/tarmac-project/tarmac/pkg/sdk/sql"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

// Tarmac provides an interface to users which wraps and simplifies the interfaces for WASM Functions execute by Tarmac. This interface
// provides access to Loggers, Metrics, KVs, and SQL Databases.
type Tarmac struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	namespace string

	// Handler registers the user function to execute as part of the Tarmac Function.
	handler func([]byte) ([]byte, error)

	// HTTP provides an interface to the Tarmac HTTP client, enabling users to make HTTP requests from functions.
	HTTP *http.Client

	// KV provides an interface to the Tarmac KVStore, enabling users to store and retrieve key-value pairs from functions.
	KV *kvstore.KV

	// SQL provides an interface to the underlying SQL datastores within Tarmac.
	SQL *sql.SQL

	// Function provides an interface to call other WASM functions registered within Tarmac.
	Function *function.Function

	// Logger provides an interface to the Tarmac structured logger, enabling users to create log messages from functions.
	Logger *logger.Logger

	// Metrics provides an interface to Tarmac metrics, enabling users to create custom metrics from functions.
	Metrics *metrics.Metrics
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
	t.handler = cfg.Handler

	// Register provided handler
	wapc.RegisterFunctions(wapc.Functions{
		"handler": t.handler,
	})

	// Set hostCall function for internal callbacks
	cfg.hostCall = wapc.HostCall

	var err error

	// Initialize a Logger instance
	t.Logger, err = logger.New(logger.Config{Namespace: cfg.Namespace, HostCall: cfg.hostCall})
	if err != nil {
		return t, fmt.Errorf("error while initializing logger - %s", err)
	}

	// Initialize a Metrics instance
	t.Metrics, err = metrics.New(metrics.Config{Namespace: cfg.Namespace, HostCall: cfg.hostCall})
	if err != nil {
		return t, fmt.Errorf("error while initializing metrics - %s", err)
	}

	// Initialize an HTTP instance
	t.HTTP, err = http.New(http.Config{Namespace: cfg.Namespace, HostCall: cfg.hostCall})
	if err != nil {
		return t, fmt.Errorf("error while initializing HTTP - %s", err)
	}

	// Initialize a KV instance
	t.KV, err = kvstore.New(kvstore.Config{Namespace: cfg.Namespace, HostCall: cfg.hostCall})
	if err != nil {
		return t, fmt.Errorf("error while initializing KV - %s", err)
	}

	// Initialize an SQL instance
	t.SQL, err = sql.New(sql.Config{Namespace: cfg.Namespace, HostCall: cfg.hostCall})
	if err != nil {
		return t, fmt.Errorf("error while initializing SQL - %s", err)
	}

	// Initialize a Function instance
	t.Function, err = function.New(function.Config{Namespace: cfg.Namespace, HostCall: cfg.hostCall})
	if err != nil {
		return t, fmt.Errorf("error while initializing Function - %s", err)
	}

	return t, nil
}
