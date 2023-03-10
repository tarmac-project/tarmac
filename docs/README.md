# Tarmac

![](tarmac-logo.png)

Tarmac: Building Serverless Applications with WebAssembly, Simplified

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac)
[![Documentation](https://img.shields.io/badge/docs-latest-blue)](https://tarmac.gitbook.io/tarmac/)
[![Build Status](https://github.com/madflojo/tarmac/actions/workflows/build.yml/badge.svg)](https://github.com/madflojo/tarmac/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac)
[![Coverage Status](https://coveralls.io/repos/github/madflojo/tarmac/badge.svg?branch=master)](https://coveralls.io/github/madflojo/tarmac?branch=master)

Tarmac is an open-source platform for building serverless applications using WebAssembly and WASI. Unlike traditional serverless platforms, Tarmac eliminates cold start times by loading functions at startup, providing near-instantaneous response times for function invocations.

Tarmac also offers a unique approach to interacting with external systems, using host callbacks to provide access to key-value stores, databases, and HTTP APIs. This makes it easy to build complex, distributed applications that can scale to meet demand.

By leveraging the security and portability of WebAssembly and WASI, Tarmac provides a lightweight and secure runtime environment for serverless functions. This enables developers to build and deploy applications quickly and easily, without worrying about the infrastructure or server maintenance.

Overall, Tarmac offers a new and innovative approach to building serverless applications, with unique features and benefits that set it apart from traditional serverless platforms.

## Tarmac vs. Serverless Functions

Tarmac shares many traits with Serverless Functions and Functions as a Service \(FaaS\) platforms. Tarmac makes it easy for developers to deploy functions and microservices without writing repetitive boilerplate code. As a developer, you can create a production-ready service in less than 100 lines of code.

But Tarmac takes Serverless Functions further. In general, FaaS platforms provide a simple runtime for user code. If a function requires any dependency \(i.e., a Database\), the developer-provided function code must maintain the database connectivity and query calls.

Using the power of WebAssembly, Tarmac not only provides functions a secure sandboxed runtime environment, but it also provides abstractions that developers can use to interact with platform capabilities such as Databases, Caching, and even Metrics.

In many ways, Tarmac is more akin to a microservices framework with the developer experience of a FaaS platform.

## Quick Start

To start executing WASM functions with Tarmac, we must define a single "Handler" function. This function will accept a byte slice as its input. The contents of this byte slice will be the HTTP payload sent to the service.

To see more examples look at one of our simple examples \(found in [example/](https://github.com/madflojo/tarmac/blob/main/example/tac/README.md)\).

```go
// Tac is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload in reverse.
package main

import (
        "fmt"
        wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
        // Tarmac uses waPC to facilitate WASM module execution. Modules must register their custom handlers
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

