## Makefile for Go example for Tarmac WASM functions

build:
	## Run TinyGo build via Docker because its easier
	tinygo build -o tarmac.wasm -target wasi main.go

docker-compose:
	docker compose up

run: build docker-compose
