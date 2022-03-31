// Echo is a small, simple Rust program that is an example WASM module for Tarmac.
// This program will accept a Tarmac server request, log it, and echo back the payload.
extern crate wapc_guest as guest;
use guest::prelude::*;

fn main() {}

#[no_mangle]
pub extern "C" fn wapc_init() {
  // Add Handler for the POST request
  register_function("POST", handler);
  // Add Handler for the PUT request
  register_function("PUT", handler);
}

fn handler(msg: &[u8]) -> CallResult {
  // Perform a host callback to log the incoming request
  let _res = host_call("tarmac", "logger", "trace", &msg.to_vec());
  Ok(msg.to_vec())
}
