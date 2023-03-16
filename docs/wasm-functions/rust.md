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
use guest::prelude::*;

fn main() {}

#[no_mangle]
pub extern "C" fn wapc_init() {}
```

Along with the waPC imports, you should also see a `wapc_init()` function created. This function is the primary entry point for Tarmac execution. We will register our handler function for Tarmac to execute using the `register_function()` function within this function.

```rust
#[no_mangle]
pub extern "C" fn wapc_init() {
  register_function("handler", handler);
}
```

In the example above, we have registered the `handler()` function. When Tarmac receives an HTTP POST request for this WASM function, it will execute the handler function as defined.

With our handler function now registered, we must create a basic version of this handler for Tarmac to call.

```rust
fn handler(msg: &[u8]) -> CallResult {}
```

As we can see from the example above, the handler input a slice of 8-bit unsigned integers, which is the raw HTTP payload. And a return value of CallResult.

## Adding Logic

Now that we have the basic structure of our WASM function created, we can start adding logic to the function and process our request.

#### Host Callbacks

One of the unique benefits of Tarmac is the ability for WASM functions to perform host callbacks to the Tarmac service itself. These Host Callbacks give users the ability to execute common framework code provided to the WASM function by Tarmac. These common framework functions can include storing data within a database, calling a remote API, or logging data. 

For our example, we will use the Host Callbacks to create a Trace log entry.

```rust
  // Perform a host callback to log the incoming request
  let _res = host_call("tarmac", "logger", "trace", &msg.to_vec());
```

For a full list of Host Callbacks checkout the [Callbacks](../callback-functions/callbacks.md) documentation.

### Do Work and Generate a Response

We can add our logic to the example, which in this case will just return the input payload.

```rust
  Ok(msg.to_vec())
```

## Full WASM function

For quick reference, the below code is the full WASM function from this example.

```rust
// Echo is a small, simple Rust program that is an example WASM module for Tarmac.
// This program will accept a Tarmac server request, log it, and echo back the payload.
extern crate wapc_guest as guest;
use guest::prelude::*;

fn main() {}

#[no_mangle]
pub extern "C" fn wapc_init() {
  register_function("handler", handler);
}

fn handler(msg: &[u8]) -> CallResult {
  // Perform a host callback to log the incoming request
  let _res = host_call("tarmac", "logger", "trace", &msg.to_vec());
  Ok(msg.to_vec())
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
name = "echo"
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
$ cp target/wasm32-unknown-unknown/release/echo.wasm functions/tarmac.wasm
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

