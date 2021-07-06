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
// This function will accept the server request, log it, and echo back the payload.
fn handler(msg: &[u8]) -> CallResult {
  // Perform a host callback to log the incoming request
  let _res = host_call("tarmac", "logger", "debug", &msg.to_vec());

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
