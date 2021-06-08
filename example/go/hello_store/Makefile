build: tests build-wasm

build-wasm:
	mkdir -p module/
	tinygo build -o module/tarmac_module.wasm -target wasi ./main.go

tests:
	mkdir -p coverage
	go test -v -covermode=count -coverprofile=coverage/coverage.out ./...
