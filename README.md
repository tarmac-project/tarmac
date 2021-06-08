# Tarmac - Microservice Framework for WASM

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac)

Tarmac is a framework for building distributed services for any language. Like many other distributed service frameworks or microservice toolkits, Tarmac abstracts the complexities of building distributed systems. Except unlike other toolkits, Tarmac is language agnostic.

Using Web Assembly (WASM), Tarmac users can write their logic in many different languages like Rust, Go, Javascript, or even C and run it using the same framework.

## Tarmac vs. Functions as a Service

Like other FaaS services, it's easy to use Tarmac to create either Functions or full Microservices. However, unlike FaaS platforms, as Web Assembly (WASM) & Web Assembly System Interface (WASI) matures, Tarmac will provide users with much more than an easy way to run a function inside a Docker container.

By leveraging Web Assembly System Interface (WASI), Tarmac creates an isolated environment for running functions. But unlike FaaS platforms, Tarmac users will be able to import Tarmac functions (still a work in progress, experimenting with gRPC currently).

Like any other microservice framework, the goal is Tarmac will handle the complexities of Database Connections, Caching, Metrics, and Dynamic Configuration. Users can focus purely on the function logic and writing it in their favorite programming language.

Tarmac aims to enable users to create robust and performant distributed services with the ease of writing serverless functions with the convenience of a standard microservices framework.

## Not ready for Production

At the moment, Tarmac is Experimental, and interfaces will change; new features will come. But if you are interested in WASM and want to write a simple Microservice, Tarmac will work.

Of course, Contributions are also welcome.

## Getting Started with Tarmac

At the moment, Tarmac can run WASI-generated WASM code. However, the serverless function code must follow a pre-defined function signature with Tarmac executing an `HTTPHandler()` function.

To understand this better, take a look at our simple example written in Go (found in [example/go](example/go)).

```golang
package main

import (
        "encoding/base64"
        "fmt"
        "os"
)

//export HTTPHandler
func HTTPHandler() int {
        d, err := base64.StdEncoding.DecodeString(os.Getenv("HTTP_PAYLOAD"))
        if err != nil {
                fmt.Fprintf(os.Stderr, "Invalid Payload")
                return 400
        }
        if os.Getenv("HTTP_METHOD") == "POST" || os.Getenv("HTTP_METHOD") == "PUT" {
                fmt.Fprintf(os.Stdout, "%s", d)
        }
        return 200
}

func main() {}
```

As we can see from the above code, Tarmac passes the HTTP Context and Payload through Environment Variables (at the moment). The Payload is Base64 encoded but otherwise untouched.

To compile the example above, run:

```shell
$ tinygo build -o tarmac_module.wasm -target wasi ./main.go
```

Once compiled, users can run Tarmac using the following command:

```shell
$ docker run -p 8443:8443 -v /path/to/certs:/certs -v ./path/to/wasm-module:/module madflojo/tarmac
```

Once running you can call the Tarmac service via `curl -v https://localhost:8443/`.
