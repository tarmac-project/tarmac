# Tarmac

![Tarmac Banner](tarmac-banner.png)

[![PkgGoDev](https://pkg.go.dev/badge/github.com/tarmac-project/tarmac/pkg/sdk)](https://pkg.go.dev/github.com/tarmac-project/tarmac/pkg/sdk)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](https://tarmac.gitbook.io/tarmac/)
[![Build Status](https://github.com/tarmac-project/tarmac/actions/workflows/build.yml/badge.svg)](https://github.com/tarmac-project/tarmac/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tarmac-project/tarmac)](https://goreportcard.com/report/github.com/tarmac-project/tarmac)
[![codecov](https://codecov.io/gh/tarmac-project/tarmac/graph/badge.svg?token=15WYYOWVCE)](https://codecov.io/gh/tarmac-project/tarmac)

## Framework for writing functions, microservices, or monoliths with Web Assembly

Tarmac is a new approach to application frameworks. Tarmac is language agnostic and offers built-in support for key/value stores like BoltDB, Redis, and Cassandra, traditional SQL databases like MySQL and Postgres, and fundamental capabilities like mutual TLS authentication and observability.

Supporting languages like Go, Rust, & Zig, you can focus on writing your functions in whatever language you like while benefiting from a robust suite of capabilities for building modern distributed services.

## Quick Start

Tarmac makes it easy to get started with building complex functions. The below function (written in Go) is an excellent example of its simplicity.

```go
// Tac is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload in reverse.
package main

import (
	"fmt"
	"github.com/tarmac-project/tarmac/pkg/sdk"
)

var tarmac *sdk.Tarmac

func main() {
	var err error

	// Initialize the Tarmac SDK
	tarmac, err = sdk.New(sdk.Config{Handler: Handler})
	if err != nil {
		return
	}
}

// Handler is the custom Tarmac Handler function that will receive a payload and
// must return a payload along with a nil error.
func Handler(payload []byte) ([]byte, error) {
	var err error

	// Log it
	tarmac.Logger.Trace(fmt.Sprintf("Reversing Payload: %s", payload))

	// Check Cache
	key := string(payload)
	rsp, err := tarmac.KV.Get(key)
	if err != nil || len(payload) < 1 {
		// Flip it and reverse
		if len(payload) > 0 {
			for i, n := 0, len(payload)-1; i < n; i, n = i+1, n-1 {
				payload[i], payload[n] = payload[n], payload[i]
			}
		}
		rsp = payload

		// Store in Cache
		err = tarmac.KV.Set(key, payload)
		if err != nil {
			tarmac.Logger.Error(fmt.Sprintf("Unable to cache reversed payload: %s", err))
			return rsp, nil
		}
	}

	// Return the payload
	return rsp, nil
}
```

To start running this function, navigate to our examples directory and run the `make build` command. The `make build` command compiles the code and generates a WebAssembly module.

```text
$ cd example/tac/go
$ make build
```

Once compiled, you can run this function as a standalone microservice using the following Docker command.

```text
$ docker run -p 8080:8080 \
  -e "APP_ENABLE_TLS=false" -e "APP_LISTEN_ADDR=0.0.0.0:8080" \
  -v `pwd`/functions:/functions madflojo/tarmac
```

With Tarmac now running, we can access our WASM function using any HTTP Client such as `curl`.

```text
$ curl -v --data "Tarmac Example" http://localhost:8080
```

That's it! You can write and deploy functions in Go, Rust, AssemblyScript, Swift, or Zig with Tarmac. For more advanced functions, check out our [developer guides](https://tarmac.gitbook.io/tarmac/wasm-functions/go).

## Multi-Function Services

While users of Tarmac can build standalone microservices with a single function quickly, it shines with multi-function services. Tarmac's ability to run multiple functions means you can create purpose-built platforms with the developer experience of serverless functions.

To get started with multi-function services, you must provide a `tarmac.json` configuration file (via the `WASM_FUNCTION_CONFIG` configuration parameter) that lists the Functions to load and the various protocols and routes to expose as endpoints. Below is a sample `tarmac.json` configuration file.

```json
{
  "services": {
    "my-service": {
      "name": "my-service",
      "functions": {
        "function1": {
          "filepath": "/path/to/function1.wasm"
        },
        "function2": {
          "filepath": "/path/to/function2.wasm"
        }
      },
      "routes": [
        {
          "type": "http",
          "path": "/function1",
          "methods": ["GET"],
          "function": "function1"
        },
        {
          "type": "http",
          "path": "/function2",
          "methods": ["POST"],
          "function": "function2"
        }
      ]
    }
  }
}
```

Each function has its own code base but shares the same service namespace and configurations in a multi-function service configuration.

In the example above, we have a service named `my-service` with `function1` and `function2` functions. Each function has a `.wasm` file at `/path/to/function1.wasm` and `/path/to/function2.wasm`.

To define the routes for each function, add a route object to the routes array with the type set to `http` and the `function` set to the function's name.

In addition to the `http` route type, Tarmac also supports `scheduled_task` routes that execute a function at a specific interval. The frequency parameter specifies the interval (in seconds).

```json
{
  "type": "scheduled_task",
  "function": "function1",
  "frequency": 10
}
```

With Tarmac's support for multiple functions, you can quickly build complex, distributed services by dividing your service into smaller, more manageable pieces.

## Architecture

Tarmac is a serverless platform that enables users to define and execute WebAssembly Functions. When Tarmac receives requests, it forwards them to WebAssembly Functions, which act as request handlers. The communication between Tarmac and WebAssembly Functions is via WebAssembly Procedure Calls (waPC).

By leveraging waPC, WebAssembly Functions can interact with Tarmac's core capabilities. Capabilities include performing callbacks to the Tarmac server to access key-value stores, interact with SQL databases, or make HTTP requests to downstream services.

To provide a streamlined developer experience, Tarmac offers a Go SDK that simplifies the usage of waPC. The SDK abstracts away the complexity of using waPC, allowing developers to focus on writing their functions and leveraging Tarmac's features.

### Example Application Architecture

The below diagram shows the architecture of an [example application](https://github.com/tarmac-project/example-airport-lookup-go/tree/main). This application demonstrates how to build a multi-function service with Tarmac using Go.

This example application will execute WebAssembly functions on boot and via a scheduler to manage airport data. The application also includes an HTTP server that serves the airport data to clients via a WebAssembly function.

```text
          +-------------------------------------------------------------------------------------------------------+                                 
          | Tarmac Host                                                                                           |                                 
          |                                       +------------------------------------------------------------+  |                                 
          |                                       | WebAssembly Engine                                         |  |                                 
          |                                       |                                                            |  |                                 
          |  +------------------------+           |  +-----------------------------------+                     |  |                                 
          |  |On Boot Function Trigger+-----------+-->Init: Creates DB Tables, Calls Load|                     |  |                                 
          |  +------------------------+           |  +--+--------------------------------+                     |  |                                 
          |                                       |     |                                                      |  |                                 
          |  +--------------------------+         |  +--v---------------------------------------------------+  |  |                                 
          |  |Scheduled Function Trigger+---------+--> Load: Calls Fetch, then loads results to SQL Database|  |  |                                 
          |  +--------------------------+         |  +--+---------------------------------------------------+  |  |                                 
          |                                       |     |                                                      |  |                                 
          |                                       |  +--v-----------------------------+                        |  |  +-----------------------------+
          |                                       |  | Fetch: Download AirportData.csv+------------------------+--+-->HTTP Server: AirportData.csv |
          |                                       |  +--------------------------------+                        |  |  +-----------------------------+
          |                                       |                                                            |  |                                 
+------+  |  +--------------------+               |  +----------------------------------+                      |  |                                 
|Client+--+-->HTTP Request Handler+---------------+-->Lookup: Fetches Data from Cache/DB|                      |  |                                 
+------+  |  +--------------------+               |  +----------------------------------+                      |  |                                 
          |                                       |                                                            |  |                                 
          |                                       +----------------------------+-------------------------------+  |                                 
          |                                                                    |                                  |                                 
          |                                                                    |                                  |                                 
          |                                                                    |                                  |                                 
          |                                       +----------------------------v-------------------------------+  |                                 
          |                                       | Tarmac Capabilities                                        |  |                                 
          |                                       |                                                            |  |                                 
          |                                       | +--------+ +------------+ +-------+ +------+               |  |                                 
          |                                       | |KV Store| |SQL Database| |Metrics| |Logger|               |  |                                 
          |                                       | +--------+ +------------+ +-------+ +------+               |  |                                 
          |                                       |                                                            |  |                                 
          |                                       +------------------------------------------------------------+  |                                 
          |                                                                                                       |                                 
          +-------------------------------------------------+-----------------------------------------------------+                                 
                                                            |                                                                                       
          +-------------------------------------------------v-----------------------------------------------------+                                 
          | External Services (Not all used in Example Application)                                               |                                 
          |                                                                                                       |                                 
          | +------+ +----------+ +-----+ +-----+ +----------+ +---------+                                        |                                 
          | |Consul| |Prometheus| |Redis| |MySQL| |PostgreSQL| |Cassandra|                                        |                                 
          | +------+ +----------+ +-----+ +-----+ +----------+ +---------+                                        |                                 
          |                                                                                                       |                                 
          +-------------------------------------------------------------------------------------------------------+                                 
```


| Language | waPC Client | Tarmac SDK |
| :--- | :--- | :--- |
| AssemblyScript | ✅ | |
| Go | ✅ | ✅ |
| Rust | ✅ | |
| Swift | ✅ | |
| Zig | ✅ | |

## Contributing

We are thrilled that you are interested in contributing to Tarmac and helping to make it even better! To get started, please check out our contributing guide for information on how to submit bug reports, feature requests, and code contributions.

### Project Contributors

<a href="https://github.com/tarmac-project/tarmac/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=tarmac-project/tarmac" />
</a>
