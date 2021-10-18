---
description: Using Tarmac as a toolkit
---

# Toolkit

In addition to being a Framework for building WASM-based distributed services, Tarmac can also be used as a toolkit to add WASM execution and host callback capabilities to existing Go applications. 

This document will show an example of importing the core capabilities of Tarmac to provide WASM host and callback capabilities.

There are two critical components to the Tarmac toolkit. The first is the WASM engine powered by waPC and thus requires guests to follow the waPC standards. The second is the Callbacks router and functionality; these provide the capabilities supplied to WASM guests and follow Tarmacs callback functions interfaces.

## WASM Guest

We will first show a simple WASM guest module that registers the `Example()` handler under the `example` function path to get started.

```go
/*
This example WASM module shows how Users can add tarmac Callback capabilities to any Go host.
*/
package main

import (
        "fmt"
        wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
        // Register the Example function for execution. Multiple functions can be registered with different call names.
        wapc.RegisterFunctions(wapc.Functions{
                "example": Example,
        })
}

// Example is a simple function that adheres to the wapc signature.
func Example(payload []byte) ([]byte, error) {
        // Execute Host Callback to log
        _, err := wapc.HostCall("tarmac", "logger", "info", payload)
        if err != nil {
                return []byte(""), fmt.Errorf("Failure - %s", err)
        }
        return []byte("Success"), nil
}
```

This WASM guest will be the guest our example application executes. The callback included is a simple logger function; however, more complex callbacks can be adopted.

## Host Example

Next, we will show a simple application that executes our WASM guest.

```go
/*
This example application shows how Tarmac can be a toolkit for adding WASM capabilities and host callback capabilities
to any Go application.
*/
package main

import (
        // Import Tarmac Callbacks Router and Desired Callback Capabilities
        "github.com/madflojo/tarmac/callbacks"
        "github.com/madflojo/tarmac/callbacks/logging"
        // Import Tarmac WASM Engine
        "github.com/madflojo/tarmac/wasm"
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
```

The example creates the Callback router before loading the WASM engine, and desired functions must also be loaded and registered before starting the WASM engine.

Each Callback capability is unique in its configuration; however, documentation is available via [package documentation](https://pkg.go.dev/github.com/madflojo/tarmac/callbacks).

We have a fully working WASM host that can run WASM guests and extend host callback capabilities with the above. For more complex examples, check out the [examples repository](https://github.com/madflojo/tarmac/tree/master/example).
