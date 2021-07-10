---
description: Creating a WASM Function in Rust.
---

# Rust

Web Assembly \(WASM\) support is a first-class feature in Rust, making Rust an excellent language to write WASM functions for Tarmac.

This guide will walk users through creating a WASM function for Tarmac in the Rust language. This walkthrough assumes there is some familiarity with the Rust tooling and language to start.

## Creating the WASM Function

We will need to begin with a new project folder, creating the `src/` directory. Within that directory, we will make our `main.rs` file.

```rust
fn main() {}
```

Tarmac internally uses a Web Assembly Procedure Calls \(waPC\) runtime, which means all WASM functions running within Tarmac must import and use a waPC compliant library.

```rust
extern crate wapc_guest as guest;
extern crate base64;
use serde::{Deserialize, Serialize};
use serde_json;
use std::collections::HashMap;
use guest::prelude::*;

fn main() {}

#[no_mangle]
pub extern "C" fn wapc_init() {}
```

You will see other imports such as `serde` and `HashMap` in the code example above. These provide capabilities for JSON parsing which will be a vital aspect later as we progress.

Along with the waPC imports, you should also see a `wapc_init()` function created. This function is the primary entry point for Tarmac execution. We will register other handler functions for Tarmac to execute using the `register_function()` function within this function.

```rust
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
```

In the example above, we have registered the `handler()` function under two Tarmac routes; `http:POST` and `http:PUT`. When Tarmac receives an HTTP POST request for this WASM function, it will execute the handler function as defined. If we wanted this function to be used for HTTP GET requests, we could add another line registering it under `http:GET`.

With our handler function now registered, we must create a basic version of this handler for Tarmac to call.

```rust
fn handler(msg: &[u8]) -> CallResult {}
```

As we can see from the example above, the handler input a slice of 8-bit unsigned integers, a raw Server Request JSON. And a return value of CallResult, which expects a Server Reply JSON back. The Server Request and Server Reply JSON are essential as they enable the Tarmac server to provide the WASM function with request context and the WASM function to provide Tarmac with how to handle the request. To understand more about Server Request and Server Response, refer to our [Inputs & Outputs](inputs-and-outputs.md) documentation.

## Adding Logic

Now that we have the basic structure of our WASM function created, we can start adding logic to the function and process our request.

### Parsing the Server Request

The first step in our logic will be to Prase the request payload. As this request payload will come in the form of a JSON, we must also define a `struct` to use to parse the JSON.

```rust
#[derive(Serialize, Deserialize)]
struct ServerRequest {
  headers: HashMap<String, String>,
  payload: String,
}

fn handler(msg: &[u8]) -> CallResult {
  // Unmarshal the request
  let rq: ServerRequest = serde_json::from_slice(msg).unwrap();
}
```

### Decoding the HTTP Payload

After parsing the Server Request JSON, the next step we need to perform is decoding the payload.

```rust
  // Decode Payload
  let b = base64::decode(rq.payload).unwrap();
```

To avoid JSON parsing conflicts, the original HTTP payload is base64 encoded. To access the original contents, we must decode them.

### Do Work and Generate a Response

We can add our logic to the example, which in this case will be a payload reverser.

```rust
  // Convert to a String
  let s = String::from_utf8(b).expect("Found Invalid UTF-8")
  let s = s.chars().rev().collect::<String>();
```

Now with our WASM function complete, we must create a Server Response JSON with our reply payload. This step will include needing to create another `struct`.

```rust
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

fn handler(msg: &[u8]) -> CallResult {
  // Unmarshal the request
  let rq: ServerRequest = serde_json::from_slice(msg).unwrap();

  // Convert to a String
  let s = String::from_utf8(b).expect("Found Invalid UTF-8")
  let s = s.chars().rev().collect::<String>();
  let enc = base64::encode(s);

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
```

Just as the incoming HTTP payload comes in base64 encoded, the Server Response must also be base64 encoded as depicted above.

### Host Callbacks

One of the unique benefits of Tarmac is the ability for WASM functions to perform host callbacks to the Tarmac service itself. These Host Callbacks give users the ability to execute standard framework code provided to the WASM function by Tarmac. These framework functions can include storing data within a database, calling a remote API, or logging data.

We will use the Host Callback to create a Trace log entry of the incoming Server Request for our example.

```rust
  // Perform a host callback to log the incoming request
  let _res = host_call("tarmac", "logger", "trace", &msg.to_vec());
```

The [Callbacks](../callback-functions/callbacks.md) documentation section explains each available Host Callback and the various actions available to WASM functions.

## Full WASM function

For quick reference, the below code is the full WASM function from this example.

```rust
// Tac is a small, simple Rust program that is an example WASM module for Tarmac.
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

// handler is a simple example of a Tarmac WASM module written in Rust.
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

## Building the WASM Function

Now that our function is ready, we must compile our Rust code into a `.wasm` file. To do this, we will need to create our Cargo manifest and build the project.

```text
$ cargo init
```

Within the `Cargo.toml` file, we must specify the different packages used in our WASM function.

```text
[package]
name = "hello"
version = "0.1.0"
authors = ["Example Developer <developer@example.com>"]
edition = "2018"

# See more keys and their definitions at https://doc.rust-lang.org/cargo/reference/manifest.html

[dependencies]
wapc-guest = "0.4.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
base64 = "0.13.0"
```

With our manifest defined, we can now build our module.

```text
$ cargo build --target wasm32-unknown-unknown --release
```

After the code build completes, we will copy the `.wasm` file into a directory Tarmac can use to run.

```text
$ mkdir -p functions
$ cp target/wasm32-unknown-unknown/release/hello.wasm functions/tarmac.wasm
```

## Running the WASM Function

We are now ready to run our WASM function via Tarmac. To make this process easier, we will be using Docker to execute Tarmac. It is not necessary to use Docker with Tarmac as it can run outside of Docker as well.

```text
$ docker run -p 8080:8080 \
  -e "APP_ENABLE_TLS=false" -e "APP_LISTEN_ADDR=0.0.0.0:8080" \
  -v ./functions:/functions madflojo/tarmac
```

In the above command, we pass two environment variables to the container using the -e flag. These environment variables will tell Tarmac to use HTTP rather than HTTPS, which is the default. For additional configuration options, check out the [Configuration](../running-tarmac/configuration.md) documentation.

With Tarmac now running, we can access our WASM function using any HTTP Client such as `curl`.

```text
$ curl -v --data "Tarmac Example" http://localhost:8080
```

## Conclusion

Developers can use this guide to get started with WASM functions and using Tarmac. Some of the information in this guide is subject to change as WASM advances. However, the concepts should stay pretty consistent.

