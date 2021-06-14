extern crate wapc_guest as guest;

use guest::prelude::*;

fn main() {}

#[no_mangle]
pub extern "C" fn wapc_init() {
  register_function("request:handler", hello_world);
}

fn hello_world(msg: &[u8]) -> CallResult {
    let _res = host_call("tarmac", "logging", "Debug", &msg.to_vec())?;
    Ok(msg.to_vec())
}
