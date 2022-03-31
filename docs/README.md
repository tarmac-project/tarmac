# Tarmac

![](tarmac-logo.png)

Framework for building distributed services with Web Assembly

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](https://tarmac.gitbook.io/tarmac/)
[![Build Status](https://github.com/madflojo/tarmac/actions/workflows/build.yml/badge.svg)](https://github.com/madflojo/tarmac/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac)
[![Coverage Status](https://coveralls.io/repos/github/madflojo/tarmac/badge.svg?branch=master)](https://coveralls.io/github/madflojo/tarmac?branch=master)


Tarmac is a unique framework designed for the next generation of distributed systems. At its core, like many other microservice frameworks, Tarmac is focused on abstracting the complexities of building cloud-native services allowing users to focus more on business logic and less on boilerplate code.

What makes Tarmac unique is that, unlike most microservice frameworks, Tarmac is language agnostic. Using Web Assembly \(WASM\), Tarmac users can write their business logic in many different languages such as Rust, Go, Javascript, or even Swift; and run it all using the same core framework.

## Tarmac vs. Serverless Functions

Tarmac shares many traits with Serverless Functions and Functions as a Service \(FaaS\) platforms. Tarmac makes it easy for developers to deploy functions and microservices without writing repetitive boilerplate code. As a developer, you can create a production-ready service in less than 100 lines of code.

But Tarmac takes Serverless Functions further. In general, FaaS platforms provide a simple runtime for user code. If a function requires any dependency \(i.e., a Database\), the developer-provided function code must maintain the database connectivity and query calls.

Using the power of Web Assembly, Tarmac not only provides functions a secure sandboxed runtime environment, but it also provides abstractions that developers can use to interact with platform capabilities such as Databases, Caching, Metrics, and even Dynamic Configuration.

In many ways, Tarmac is more akin to a microservices framework with the developer experience of a FaaS platform.

## Quick Start

At the moment, Tramac is executing WASM functions by executing a defined set of function signatures. When Tarmac receives an HTTP GET request, it will call the function's registered under the `GET` signature.

As part of the WASM Function, users must register their handlers using the pre-defined function signatures.

To understand this better, look at one of our simple examples \(found in [example/](https://github.com/madflojo/tarmac/blob/main/example/tac/README.md)\).

```golang
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
                // Register a GET request handler
                "GET": NoHandler,
                // Register a POST request handler
                "POST": Handler,
                // Register a PUT request handler
                "PUT": Handler,
                // Register a DELETE request handler
                "DELETE": NoHandler,
        })
}

// NoHandler is a custom Tarmac Handler function that will return an error that denies
// the client request.
func NoHandler(payload []byte) ([]byte, error) {
        return []byte(""), fmt.Errorf("Not Implemented")
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

Tarmac passes the HTTP Payload to the WASM function untouched.

To compile the example above, run:

```text
$ cd example/tac/go
$ make build
```

Once compiled, users can run Tarmac via Docker using the following command:

```text
$ docker run -p 8080:8080 \
  -e "APP_ENABLE_TLS=false" -e "APP_LISTEN_ADDR=0.0.0.0:8080" \
  -v ./functions:/functions madflojo/tarmac
```

With Tarmac now running, we can access our WASM function using any HTTP Client such as `curl`.

```text
$ curl -v --data "Tarmac Example" http://localhost:8080
```

