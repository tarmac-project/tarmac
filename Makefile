# Makefile is used to drive the build and installation of this application
# this is meant to be used with a local copy of code repository.

build: build-testdata

build-testdata:
	$(MAKE) -C testdata/default build
	$(MAKE) -C testdata/kv build
	$(MAKE) -C testdata/sql build
	$(MAKE) -C testdata/logger build

tests: build
	@echo "Launching Tests in Docker Compose"
	docker-compose -f dev-compose.yml up -d cassandra-primary cassandra mysql consul consulator
	sleep 120 
	docker-compose -f dev-compose.yml up --exit-code-from tests --build tests

benchmarks:
	@echo "Launching Tests in Docker Compose"
	docker-compose -f dev-compose.yml up -d cassandra-primary cassandra mysql
	sleep 120
	docker-compose -f dev-compose.yml up --build benchmarks

clean:
	@echo "Cleaning up build junk"
	-docker-compose -f dev-compose.yml down

tarmac:
	@echo "Starting Application"
	docker-compose -f dev-compose.yml up --build tarmac

tarmac-performance: build
	@echo "Starting Application"
	docker-compose -f dev-compose.yml up -d tarmac-performance
