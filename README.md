# Tarmac - Microservice Framework for WASM

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac)

Tarmac is a framework for building distributed services for any language. Like many other distributed service frameworks or microservice toolkits, Tarmac abstracts the complexities of building distributed systems. Except unlike other toolkits, Tarmac is language agnostic.

Using Web Assembly (WASM), Tarmac users can write their logic in many different languages like Rust, Go, Javascript, or even C and run it using the same framework.

## Tarmac vs. Functions as a Service

Like other FaaS services, it's easy to use Tarmac to create either Functions or full Microservices. However, unlike FaaS platforms, as Web Assembly (WASM) & Web Assembly System Interface (WASI) matures, Tarmac will provide users with much more than an easy way to run a function inside a Docker container.

By leveraging Web Assembly System Interface (WASI), Tarmac creates an isolated environment for running functions. But unlike FaaS platforms, Tarmac users will be able to import Tarmac functions (still a work in progress).

Like any other microservice framework, the goal is Tarmac will handle the complexities of Database Connections, Caching, Metrics, and Dynamic Configuration. Users can focus purely on the function logic and writing it in their favorite programming language.

Tarmac aims to enable users to create robust and performant distributed services with the ease of writing serverless functions with the convenience of a standard microservices framework.

## Not ready for Production

At the moment, Tarmac is Experimental, and interfaces will change; new features will come. But if you are interested in WASM and want to write a simple Microservice, Tarmac will work.

Of course, Contributions are also welcome.

## Getting Started with Tarmac

At the moment, Tarmac can run WASI-generated WASM code. However, the serverless function code must follow a pre-defined function signature with Tarmac executing a defined `request:handler` function.

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

Tarmac passes the HTTP Context and Payload to the WASM guest via the incoming `msg`. The `msg` is a JSON which contains Headers and a Payload which is Base64 encoded but otherwise untouched.

To compile the example above, run:

```shell
$ cd example/hello
$ make build
```

Once compiled, users can run Tarmac using the following command:

```shell
$ docker run -p 8443:8443 -v /path/to/certs:/certs -v ./path/to/wasm-module:/module madflojo/tarmac
```

Once running you can call the Tarmac service via `curl -v https://localhost:8443/`.
