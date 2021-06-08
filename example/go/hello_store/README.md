# Tarmac Go Example

WARNING: This specific module is not compilable and is a work in progress, please checkout the http_env module instead.


This project is a Go example of building a WASM microservice using Tarmac.

Tarmac is a framework for building distributed services for any language. Like many other distributed service frameworks or microservice toolkits, Tarmac abstracts the complexities of building distributed systems. Except unlike other toolkits, Tarmac is language agnostic.

Using Web Assembly (WASM), Tarmac users can write their logic in many different languages like Rust, Go, Javascript, or even C and run it using the same framework.

This project aims to show users how Tarmac can be used with Go & TinyGo to build a WASM microservice.

## Getting Started

To run this microservice with Tarmac, first, we must build a `.wasm` file from our Go code using TinyGo.

```shell
$ tinygo build -o module/tarmac_module.wasm -target wasi ./main.go
```

Once complete, we can run the following Docker command to start our microservice.
 
```shell
$ docker run -p 443:8443 -v certs:/certs -v module/:/module madflojo/tarmac
```

By default, Tarmac expects to run over HTTPS, which requires a certificate. If running this project for development or fun, you can disable this by running the following Docker command instead.

```shell
$ docker run -p 80:8080 -v module/:/module -e "APP_LISTEN_ADDR=0.0.0.0:8080" -e "APP_ENABLE_TLS=false" madflojo/tarmac
```

Or, just run `docker-compose up`
