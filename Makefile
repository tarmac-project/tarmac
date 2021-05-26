# Makefile is used to drive the build and installation of this application
# this is meant to be used with a local copy of code repository.

tests:
	@echo "Launching Tests in Docker Compose"
	mkdir -p example/go/module/
	tinygo build -o example/go/module/tarmac_module.wasm -target wasi ./example/go/main.go
	docker-compose -f dev-compose.yml up --build tests

clean:
	@echo "Cleaning up build junk"
	-docker-compose -f dev-compose.yml down

tarmac:
	@echo "Starting Application"
	docker-compose -f dev-compose.yml up --build tarmac
