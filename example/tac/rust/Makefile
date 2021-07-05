## Makefile for Rust example for Tarmac WASM modules

build:
	cargo build --target wasm32-unknown-unknown --release
	mkdir -p module
	cp target/wasm32-unknown-unknown/release/hello.wasm module/tarmac_module.wasm

docker-compose:
	docker compose up

run: build docker-compose
	
