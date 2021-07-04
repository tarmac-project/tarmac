# Tarmac Example

This project is an example of building a WASM-based microservice using Tarmac.

Tarmac is a framework for building distributed services in any language. Like many other distributed 
service/microservice frameworks, Tarmac abstracts the complexities of building distributed systems, eliminating the 
need for boilerplate code for standard functionality. Except, unlike other frameworks, Tarmac is language agnostic.

Using Web Assembly (WASM), Tarmac users can write their logic in different languages such as Rust, Go, Javascript, or 
even C and run it using the same framework.

## Getting Started

To run this microservice with Tarmac, we must first compile the code into a WASM module. To do this, execute the `make 
build` command.

```console
$ make build
```

Once we have a WASM executable, we can launch Tarmac with Docker. To simplify this, we've included an example Docker 
Compose file.

```console
$ docker compose up tarmac-example
```

Once running, you can send requests to this service using `curl`.

```console
$ curl -X POST --data "This is a test" http://localhost -v
```
