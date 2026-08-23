TINYGO_LEGACY_IMAGE ?= tinygo/tinygo:0.34.0
TINYGO_CURRENT_IMAGE ?= tinygo/tinygo:0.41.1
GOPROXY_URL ?= https://proxy.golang.org

.PHONY: build build-legacy build-current module-check

build: build-legacy build-current

build-legacy:
	docker run --rm -v `pwd`:/build -w /build -u root $(TINYGO_LEGACY_IMAGE) tinygo build -o /build/tarmac-0.34.wasm -target=wasi /build/main.go

build-current:
	docker run --rm -v `pwd`:/build -w /build -u root $(TINYGO_CURRENT_IMAGE) tinygo build -o /build/tarmac.wasm -target=wasip1 -scheduler=none --no-debug -buildmode=c-shared /build/main.go

module-check:
	GOWORK=off GOPROXY=$(GOPROXY_URL) go mod tidy -diff
	GOWORK=off GOPROXY=$(GOPROXY_URL) go mod download
	GOWORK=off GOPROXY=$(GOPROXY_URL) go list -mod=readonly ./...
