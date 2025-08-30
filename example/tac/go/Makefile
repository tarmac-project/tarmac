## Makefile for Go example for Tarmac WASM functions

build:
	mkdir -p functions
	tinygo build -o functions/tarmac.wasm -target wasi main.go

docker-compose:
	docker compose up

run: build docker-compose
