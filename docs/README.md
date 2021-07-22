# Tarmac

![](tarmac-logo.png)

Framework for building distributed services with Web Assembly

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac) [![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac) [![Documentation](https://img.shields.io/badge/Docs-latest-blue)](https://tarmac.gitbook.io/tarmac/)

Tarmac is a unique framework designed for the next generation of distributed systems. At its core, like many other microservice frameworks, Tarmac is focused on abstracting the complexities of building cloud-native services allowing users to focus more on business logic and less on boilerplate code.

What makes Tarmac unique is that, unlike most microservice frameworks, Tarmac is language agnostic. Using Web Assembly \(WASM\), Tarmac users can write their business logic in many different languages such as Rust, Go, Javascript, or even Swift; and run it all using the same core framework.

## Tarmac vs. Serverless Functions

Tarmac shares many traits with Serverless Functions and Functions as a Service \(FaaS\) platforms. Tarmac makes it easy for developers to deploy functions and microservices without writing repetitive boilerplate code. As a developer, you can create a production-ready service in less than 100 lines of code.

But Tarmac takes Serverless Functions further. In general, FaaS platforms provide a simple runtime for user code. If a function requires any dependency \(i.e., a Database\), the developer-provided function code must maintain the database connectivity and query calls.

Using the power of Web Assembly, Tarmac not only provides functions a secure sandboxed runtime environment, but it also provides abstractions that developers can use to interact with platform capabilities such as Databases, Caching, Metrics, and even Dynamic Configuration.

In many ways, Tarmac is more akin to a microservices framework with the developer experience of a FaaS platform.

## Quick Start

At the moment, Tramac is executing WASM functions by executing a defined set of function signatures. When Tarmac receives an HTTP GET request, it will call the function's registered under the `http:GET` signature.

As part of the WASM Function, users must register their handlers using the pre-defined function signatures.

To understand this better, look at one of our simple examples written in Rust \(found in [example/](https://github.com/madflojo/tarmac/tree/e1e6e952a1f6e2f89448e17d15862e199ff64e84/docs/example/README.md)\).

```rust
// Tac is a small, simple Rust program that is an example WASM function for Tarmac.
// This program will accept a Tarmac server request, log it, and echo back the payload
// but with the payload reversed.
extern crate wapc_guest as guest;
extern crate base64;
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
  // Add Handler for the GET request
  register_function("http:GET", fail_handler);
  // Add Handler for the POST request
  register_function("http:POST", handler);
  // Add Handler for the PUT request
  register_function("http:PUT", handler);
  // Add Handler for the DELETE request
  register_function("http:DELETE", fail_handler);
}

// fail_handler will accept the server request and return a server response
// which rejects the client request
fn fail_handler(_msg: &[u8]) -> CallResult {
  // Create the response
  let rsp = ServerResponse {
      status: Status {
        code: 503,
        status: "Not Implemented".to_string(),
      },
      payload: "".to_string(),
      headers: HashMap::new(),
  };

  // Marshal the response
  let r = serde_json::to_vec(&rsp).unwrap();

  // Return JSON byte array
  Ok(r)
}

// handler is a simple example of a Tarmac WASM function written in Rust.
// This function will accept the server request, log it, and echo back the payload
// but with the payload reversed.
fn handler(msg: &[u8]) -> CallResult {
  // Perform a host callback to log the incoming request
  let _res = host_call("tarmac", "logger", "debug", &msg.to_vec());

  // Unmarshal the request
  let rq: ServerRequest = serde_json::from_slice(msg).unwrap();

  // Decode Payload
  let b = base64::decode(rq.payload).unwrap();
  // Convert to a String
  let s = String::from_utf8(b).expect("Found Invalid UTF-8");
  // Reverse it and re-encode
  let enc = base64::encode(s.chars().rev().collect::<String>());

  // Create the response
  let rsp = ServerResponse {
      status: Status {
        code: 200,
        status: "OK".to_string(),
      },
      payload: enc,
      headers: HashMap::new(),
  };

  // Marshal the response
  let r = serde_json::to_vec(&rsp).unwrap();

  // Return JSON byte array
  Ok(r)
}
```

Tarmac passes the HTTP Context and Payload to the WASM function via the incoming `msg`. The `msg` is a JSON that contains Headers and a Payload which is Base64 encoded but otherwise untouched.

To compile the example above, run:

```text
$ cd example/tac/rust
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

