# Tarmac - Framework for building distributed services with Web Assembly

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac)
[![Documentation](https://img.shields.io/badge/Docs-latest-blue)](https://madflojo.gitbook.io/tarmac/)

Tarmac is a unique framework designed for the next generation of distributed systems. At its core, like many other microservice frameworks, Tarmac is focused on abstracting the complexities of building cloud-native services allowing users to focus more on business logic and less on boilerplate code. 

What makes Tarmac unique is that, unlike most microservice frameworks, Tarmac is language agnostic. Using Web Assembly (WASM), Tarmac users can write their business logic in many different languages such as Rust, Go, Javascript, or even Swift; and run it all using the same core framework.

## Tarmac vs. Serverless Functions

Tarmac shares many traits with Serverless Functions and Functions as a Service (FaaS) platforms. Tarmac makes it easy for developers to deploy functions and microservices without writing repetitive boilerplate code. As a developer, you can create a production-ready service in less than 100 lines of code.

But Tarmac takes Serverless Functions further. In general, FaaS platforms provide a simple runtime for user code. If a function requires any dependency (i.e., a Database), the developer-provided function code must maintain the database connectivity and query calls.

Using the power of Web Assembly, Tarmac not only provides functions a secure sandboxed runtime environment, but it also provides abstractions that developers can use to interact with platform capabilities such as Databases, Caching, Metrics, and even Dynamic Configuration. 

In many ways, Tarmac is more akin to a microservices framework with the developer experience of a FaaS platform.

## Not ready for Production

At the moment, Tarmac is Experimental, and interfaces will change; new features will come. But if you are interested in WASM and want to write a simple Function, Tarmac will work.

## Getting Started with Tarmac

At the moment, Tramac is executing WASM functions or "Tarmac Modules" by executing a defined set of function signatures. When Tarmac receives an HTTP GET request, it will call the function's registered under the `http:GET` signature.

As part of the Tarmac Module, users must register their functions using the pre-defined function signatures.

To understand this better, take a look at our simple example written in Rust (found in [example/](example/)).

```rust
// Hello is a small, simple Rust program that is an example WASM module for Tarmac.
// This program will accept a Tarmac server request, log it, and echo back the payload.
extern crate wapc_guest as guest;
use serde::{Deserialize, Serialize};
use serde_json;
use std::collections::HashMap;
use guest::prelude::*;

#[derive(Serialize, Deserialize)]
struct ServerRequest {
  headers: HashMap<String, String>,
  payload: String,
}

#[derive(Serialize, Deserialize)]
struct ServerResponse {
  headers: HashMap<String, String>,
  status: Status,
  payload: String,
}

#[derive(Serialize, Deserialize)]
struct Status {
  code: u32,
  status: String,
}

fn main() {}

#[no_mangle]
pub extern "C" fn wapc_init() {
  register_function("request:handler", hello_world);
}

// hello_world is a simple example of a Tarmac WASM module written in Rust.
// This function will accept the server request, log it, and echo back the payload.
fn hello_world(msg: &[u8]) -> CallResult {
  // Perform a host callback to log the incoming request
  let _res = host_call("tarmac", "logging", "debug", &msg.to_vec());

  // Unmarshal the request
  let rq: ServerRequest = serde_json::from_slice(msg).unwrap();

  // Create the response
  let rsp = ServerResponse {
      status: Status {
        code: 200,
        status: "OK".to_string(),
      },
      payload: rq.payload,
      headers: HashMap::new(),
  };

  // Marshal the response
  let r = serde_json::to_vec(&rsp).unwrap();

  // Return JSON byte array
  Ok(r)
}
```

Tarmac passes the HTTP Context and Payload to the WASM function via the incoming `msg`. The `msg` is a JSON which contains Headers and a Payload which is Base64 encoded but otherwise untouched.

To compile the example above, run:

```shell
$ cd example/tac/rust
$ make build
```

Once compiled, users can run Tarmac using the following command:

```shell
$ docker run -p 8443:8443 -v /path/to/certs:/certs -v ./module:/module madflojo/tarmac
```

Once running you can call the Tarmac service via `curl -v https://localhost:8443/`.
