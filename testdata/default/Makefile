## Makefile for Go example for Tarmac WASM functions

build:
	## Run TinyGo build via Docker because its easier
	docker run -v `pwd`/:/src -w /src tinygo/tinygo:latest tinygo build -o /src/tarmac.wasm -target wasi /src/main.go

docker-compose:
	docker compose up

run: build docker-compose
