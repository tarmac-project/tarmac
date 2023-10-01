---
description: Creating a WASM Function in Go.
---

# Go

Tarmac leverages the Web Assembly System Interface (WASI), which is currently only supported by [TinyGo](https://tinygo.org/). While TinyGo has many features, there are some limitations to its support.

However, thanks to the Go SDK, writing Tarmac functions is quick and easy. This guide will walk users through creating a simple function using Go.

## Basic WASM function

We will first need to begin with a new project folder creating a `main.go` file within it. This file will hold all of our application logic.

Within our `main.go` file; we will need first to import the Tarmac Go SDK.

```go
package main

import (
	"github.com/tarmac-project/tarmac/pkg/sdk"
)
```

Once we've imported the SDK, we will need to both create our Function and register it with the SDK. We will start by initializing the SDK and registering our Function, `Handler()`.

```go
func main() {
	// Initialize the Tarmac SDK
	_, err := sdk.New(sdk.Config{Namespace: "test-service", Handler: Handler})
	if err != nil {
		return
	}
}
```

As Tarmac receives requests such as an HTTP POST request, the `Handler()` function will be called with the HTTP payload provided as the `payload` parameter.

We can create our Function, which returns a simple "Howdie" message.

```go
func Handler(payload []byte) ([]byte, error) {
	// Return a happy message
	return []byte("Howdie"), nil
}
```

### Building the WASM Function

Now that our function is ready, we must compile our Go code into a `.wasm` file. To do this, we will be using TinyGo.

```text
$ mkdir -p functions
$ tinygo build -o functions/tarmac.wasm -target wasi main.go
```

The first step above is using `mkdir` to create a functions directory, this is not required but will be helpful when running Tarmac in the next stage.

After the functions directory is created, we are using the `tinygo` command to build our `.wasm` file. The inclusion of `-target wasi` is important as it directs TinyGo to compile the Go code using the Web Assembly System Interface \(wasi\) standard. This standard is useful for running Web Assembly on the server vs. on the browser.

With this step complete, we have built our WASM function.

### Running the WASM Function

We are now ready to run our WASM function via Tarmac. To make this process easier, we will be using Docker to execute Tarmac. It is not necessary to use Docker with Tarmac as it can run outside of Docker as well.

```text
$ docker run -p 8080:8080 \
  -e "APP_ENABLE_TLS=false" -e "APP_LISTEN_ADDR=0.0.0.0:8080" \
  -v `pwd`./functions:/functions madflojo/tarmac
```

In the above command, we are passing two environment variables to the container using the `-e` flag. These environment variables will tell Tarmac to use HTTP rather than HTTPS, which is the default. For additional configuration options, check out the [Configuration](../running-tarmac/configuration.md) documentation.

With Tarmac now running, we can access our WASM function using any HTTP Client such as `curl`.

```text
$ curl -v --data "Tarmac Example" http://localhost:8080
```

## Expanding beyond Hello World

While the above Hello World example provides an excellent introduction to creating Go functions, it does not showcase the power of Tarmac.

Tarmac provides integrations with many of the capabilities required to build today's modern platforms. These capabilities include Key:Value datastores such as Redis or Cassandra. The ability to create metrics for observability and log messages via a structured logger. Or even the ability to call HTTP end-points with an HTTP Client.

These integrations are simple to call with the Go SDK; the below Function showcases several capabilities, such as calling a key:value cache and logging.

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

### Conclusion

Developers can use this guide to get started with WASM functions and using Tarmac. Some of the information in this guide is subject to change as support for WASM in Go advances.

