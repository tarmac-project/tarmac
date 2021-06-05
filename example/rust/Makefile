build:
	@echo "Building WASM Modules"
	cargo wasi build --release
	mkdir -p module
	cp ./target/wasm32-wasi/release/rust.wasi.wasm ./module/tarmac_module.wasm

docker-compose:
	@echo "Starting Docker"
	docker compose up

run: build docker-compose
