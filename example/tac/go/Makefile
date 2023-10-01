## Makefile for Go example for Tarmac WASM functions

build:
	mkdir -p functions
	docker run -v `pwd`/:/build -w /build tinygo/tinygo:latest tinygo build -o /build/functions/tarmac.wasm -target wasi /build/main.go

docker-compose:
	docker compose up

run: build docker-compose
