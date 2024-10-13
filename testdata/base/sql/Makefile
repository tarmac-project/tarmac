## Makefile for Go example for Tarmac WASM functions

build:
	## Run TinyGo build via Docker because its easier
	docker run --rm -v `pwd`/:/tarmacbuild -w /tarmacbuild tinygo/tinygo:0.32.0 tinygo build -o /tarmacbuild/tarmac.wasm -target wasi /tarmacbuild/main.go

docker-compose:
	docker compose up

run: build docker-compose
