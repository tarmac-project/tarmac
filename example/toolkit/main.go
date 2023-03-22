/*
This example application shows how Tarmac can be a toolkit for adding WASM capabilities and host callback capabilities
to any Go application.
*/
package main

import (
	// Import Tarmac Callbacks Router and Desired Callback Capabilities
	"github.com/madflojo/tarmac/pkg/callbacks"
	"github.com/madflojo/tarmac/pkg/callbacks/logging"
	// Import Tarmac WASM Engine
	"github.com/madflojo/tarmac/pkg/wasm"
	"github.com/sirupsen/logrus"
)

func main() {

	// Create Logger instance
	logger := logrus.New()

	// Create Callback Router for Tarmac Callback functions
	router := callbacks.New(callbacks.Config{})

	// Create Callback Logging instance
	callBackLogger, err := logging.New(logging.Config{Log: logger})
	if err != nil {
		logger.Errorf("Unable to create new logger instance - %s", err)
		return
	}

	// Register Info log callback
	router.RegisterCallback("logger", "info", callBackLogger.Info)

	// Start WASM Engine
	engine, err := wasm.NewServer(wasm.Config{
		// Register our callback router
		Callback: router.Callback,
	})
	if err != nil {
		logger.Errorf("Unable to initiate WASM server instance - %s", err)
	}

	// Load WASM module
	err = engine.LoadModule(wasm.ModuleConfig{
		Name:     "example",
		Filepath: "./wasm/example.wasm",
	})
	if err != nil {
		logger.Errorf("Unable to load WASM module - %s", err)
	}

	// Fetch Module instance
	module, err := engine.Module("example")
	if err != nil {
		logger.Errorf("Unable to fetch instance of module - %s", err)
	}

	// Execute Module with custom payload
	r, err := module.Run("example", []byte("Hello"))
	if err != nil {
		logger.Errorf("Unable to execute wasm module - %s", err)
	}

	// Log results
	logger.Infof("Module execution - %s", r)
}
