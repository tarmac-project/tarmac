# Tarmac Example: KV Counter

This project is an example of building a WASM-based microservice using Tarmac. This service is a KV Counter
service that can be run both as an HTTP function or a scheduled task.

To fetch the current value of the Counter, call the service with a GET request.

```console
$ curl http://localhost
6
```

To increment the Counter, send a POST request with no payload.

```console
$ curl -X POST http://localhost
7
```

In addition to the HTTP handlers, this WASM function runs as a scheduled task. Every 30 seconds, the Counter 
will be incremented by one.

Tarmac is a framework for building distributed services in any language. Like many other distributed 
service/microservice frameworks, Tarmac abstracts the complexities of building distributed systems, eliminating the 
need for boilerplate code for standard functionality. Except, unlike other frameworks, Tarmac is language agnostic.

Using Web Assembly (WASM), Tarmac users can write their logic in different languages such as Rust, Go, Javascript, or 
even C and run it using the same framework.
