---
description: Powered by Web Assembly Procedure Calls
---

# Callbacks

Tarmac, at its core, is powered by the Web Assembly Procedure Call (waPC) Project. The Web Assembly Procedure Call Project defines imported and exported functions between a WASM host and a guest WASM function. A WASM runtime host like Tarmac and a WASM Function running within Tarmac can communicate back and forth using these functions.

A prime example of this is the `HostCall()` function used by guest WASM functions. This `HostCall()` is a callback function that enables WASM functions to pass back data to the Tarmac host for the explicit goal of executing host-level functionality.

This ability to perform a Host Callback is what sets Tarmac apart from most other serverless runtimes. A WASM function can use the Host Callback functionality to access a full suite of standard functionality that would traditionally be too heavy for a serverless function.

Essentially, Host Callbacks allow Tarmac to provide developers the functionality of a standard Microservice Framework along with the convenience of a serverless runtime.

## Using Host Callbacks

Calling a Host Callback is relatively straightforward for WASM functions. As outlined in the language guides, each WASM function must import a waPC compliant guest library. This guest library will allow users to access a `HostCall()` function for their language of choice. 

The example below is an example of calling the Host Callback function in Go.

```golang
_, err := wapc.HostCall("tarmac", "logger", "debug", payload)
```

The `HostCall()` function takes three parameters. The first is the namespace which developers should always set to `tarmac`. The second is the capability requested, such as `logger` or `kvstore`. The third is the function to execute; for a `kvstore` capability, we may want to perform a `get` or a `set`. 

This section of documentation outlines all of the various host-level capabilities Tarmac provides. Each unit will outline the capabilities, functions, and input/output data.

