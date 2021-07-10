## Makefile for Rust example for Tarmac WASM functions

build:
	cargo build --target wasm32-unknown-unknown --release
	mkdir -p functions
	cp target/wasm32-unknown-unknown/release/hello.wasm functions/tarmac.wasm

docker-compose:
	docker compose up

run: build docker-compose
	
