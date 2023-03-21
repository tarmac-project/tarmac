---
description: Creating a WASM Function in Go.
---

# Go

At the time of this writing, Web Assembly support in Go is officially listed as Experimental. However, the basics of Web Assembly do work in Go. In fact, Tarmac is written in Go and using a Go-based WASM Host Runtime.

To create a WASM Function for Tarmac, users will need to use [TinyGo](https://tinygo.org/) to compile the WASM Function into a `.wasm` file. There are some limitations on which version of TinyGo to use, and it is advised to reference the [Supported Languages](supported-languages.md) page before continuing.

### Basic WASM Function structure

We will need to begin with a new project folder, creating a `main.go` file within it. This file will hold all of our application logic.

```go
package main
```

Tarmac internally uses a Web Assembly Procedure Calls \(waPC\) runtime, which means all WASM Functions running in Tarmac must import and use a waPC compliant library.

```go
import (
	"fmt"
	wapc "github.com/wapc/wapc-guest-tinygo"
)
```

Once the waPC package is imported, we will create a `main()` function; this function will be our primary entry point for Tarmac execution. Within this function, we will register our handler function for Tarmac to execute using the `wapc.RegisterFunctions` function.

```go
func main() {
	wapc.RegisterFunctions(wapc.Functions{
		"handler": Handler,
	})
}
```

In the example above, we have registered the `Handler` function. When Tarmac receives an HTTP POST request for this WASM Function, it will execute the handler function as defined.

With our handler function registered, we must create a basic function.

```go
func Handler(payload []byte) ([]byte, error) {
	return []byte(`Success`), nil
}

```

As we can see from the example above, the handler has a byte slice input and return value. The HTTP payload is sent as the input untouched.

### Adding Logic

Now that we have the basic structure of our WASM function created, we can start adding logic to the function and process our request.

In order to avoid conflicts with the Server Request JSON, the original HTTP payload is base64 encoded. To access the original contents, we must decode them.

#### Host Callbacks

One of the unique benefits of Tarmac is the ability for WASM functions to perform host callbacks to the Tarmac service itself. These Host Callbacks give users the ability to execute common framework code provided to the WASM function by Tarmac. These common framework functions can include storing data within a database, calling a remote API, or logging data. 

For our example, we will use the Host Callbacks to create a Trace log entry.

```go
	// Perform a host callback to log the incoming request
	_, err = wapc.HostCall("tarmac", "logger", "trace", []byte(fmt.Sprintf("Reversing Payload: %s", s)))
	if err != nil {
		return []byte(fmt.Sprintf("Failed to call host callback - %s", err)), nil
	}
```

For a full list of Host Callbacks checkout the [Callbacks](../callback-functions/callbacks.md) documentation.

#### Do Work and Return a Response

We can add our logic to the example, which in this case will be a payload reverser.

```go
	// Flip it and reverse
	if len(payload) > 0 {
		for i, n := 0, len(payload)-1; i < n; i, n = i+1, n-1 {
			payload[i], payload[n] = payload[n], payload[i]
		}
	}
```

Now with our WASM function complete, we must return a response.

```go
	return payload, nil
```

### Full WASM function

For quick reference, below is the full WASM function from this example.

```go
// Tac is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload in reverse.
package main

import (
        "fmt"
        wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
        // Tarmac uses waPC to facilitate WASM module execution. Modules must register their custom handlers under the
        // appropriate method as shown below.
        wapc.RegisterFunctions(wapc.Functions{
                "handler": Handler,
        })
}

// Handler is the custom Tarmac Handler function that will receive a payload and
// must return a payload along with a nil error.
func Handler(payload []byte) ([]byte, error) {
        // Perform a host callback to log the incoming request
        _, err := wapc.HostCall("tarmac", "logger", "trace", []byte(fmt.Sprintf("Reversing Payload: %s", payload)))
        if err != nil {
                return []byte(""), fmt.Errorf("Unable to call callback - %s", err)
        }

        // Flip it and reverse
        if len(payload) > 0 {
                for i, n := 0, len(payload)-1; i < n; i, n = i+1, n-1 {
                        payload[i], payload[n] = payload[n], payload[i]
                }
        }

        // Return the payload via a ServerResponse JSON
        return payload, nil
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

### Conclusion

Developers can use this guide to get started with WASM functions and using Tarmac. Some of the information in this guide is subject to change as support for WASM in Go advances. However, the concepts of Tarmac and WASM functions should stay fairly consistent.

