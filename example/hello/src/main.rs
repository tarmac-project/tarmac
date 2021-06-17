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
