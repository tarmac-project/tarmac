## Makefile for Go example for Tarmac WASM functions

build:
	## Run TinyGo build via Docker because its easier
	docker run --rm -v `pwd`:/build -w /build -u root tinygo/tinygo:0.34.0 tinygo build -o /build/tarmac.wasm -target wasi /build/main.go

docker-compose:
	docker compose up

run: build docker-compose
