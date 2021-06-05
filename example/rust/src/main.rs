use std::env::var;
use std:: str;
use base64::{decode};


fn main() {
}

#[no_mangle]
pub extern "C" fn HTTPHandler() -> i32 {
  let payload = var("HTTP_PAYLOAD").unwrap();
  if payload.len() > 0 {
    let decoded = decode(&payload).unwrap();
    println!("{}", str::from_utf8(&decoded).unwrap());
    return 200
  }
  return 500
}

