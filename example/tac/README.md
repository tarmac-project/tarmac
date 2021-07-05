# Tarmac Example: Tac

This project is an example of building a WASM-based microservice using Tarmac. This service is a request reversal service 
that takes the user-provided HTTP Payload and returns that payload to the user but in reverse.

```console
$ curl --data "Tarmac Example" -v http://localhost
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 80 (#0)
> POST / HTTP/1.1
> Host: localhost
> User-Agent: curl/7.64.1
> Accept: */*
> Content-Length: 14
> Content-Type: application/x-www-form-urlencoded
> 
* upload completely sent off: 14 out of 14 bytes
< HTTP/1.1 200 OK
< Date: Mon, 05 Jul 2021 01:53:13 GMT
< Content-Length: 14
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host localhost left intact
elpmaxE camraT
```

Tarmac is a framework for building distributed services in any language. Like many other distributed 
service/microservice frameworks, Tarmac abstracts the complexities of building distributed systems, eliminating the 
need for boilerplate code for standard functionality. Except, unlike other frameworks, Tarmac is language agnostic.

Using Web Assembly (WASM), Tarmac users can write their logic in different languages such as Rust, Go, Javascript, or 
even C and run it using the same framework.
