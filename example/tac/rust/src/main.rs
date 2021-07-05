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
  register_function("request:handler", handler);
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
