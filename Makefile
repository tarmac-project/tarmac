# Makefile is used to drive the build and installation of this application
# this is meant to be used with a local copy of code repository.

build: build-testdata

build-testdata:
	$(MAKE) -C testdata/sdkv1/kv build
	$(MAKE) -C testdata/sdkv1/sql build
	$(MAKE) -C testdata/sdkv1/logger build
	$(MAKE) -C testdata/base/default build
	$(MAKE) -C testdata/base/fail build
	$(MAKE) -C testdata/base/kv build
	$(MAKE) -C testdata/base/sql build
	$(MAKE) -C testdata/base/logger build
	$(MAKE) -C testdata/base/function build
	$(MAKE) -C testdata/base/successafter5 build

tests: build tests-nobuild
tests-nobuild: tests-base tests-redis tests-cassandra tests-mysql tests-postgres tests-boltdb tests-inmemory

tests-base:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up -d consul consulator
	docker compose -f dev-compose.yml up --exit-code-from tests-base --build tests-base
	docker compose -f dev-compose.yml down

tests-boltdb:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up --exit-code-from tests-boltdb tests-boltdb
	docker compose -f dev-compose.yml down

tests-inmemory:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up --exit-code-from tests-inmemory tests-inmemory
	docker compose -f dev-compose.yml down

tests-redis:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up --exit-code-from tests-redis tests-redis
	docker compose -f dev-compose.yml down

tests-cassandra:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up -d cassandra-primary cassandra
	docker compose -f dev-compose.yml up --exit-code-from tests-cassandra tests-cassandra
	docker compose -f dev-compose.yml down

tests-mysql:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up -d mysql
	docker compose -f dev-compose.yml up --exit-code-from tests-mysql tests-mysql
	docker compose -f dev-compose.yml down

tests-postgres:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up -d postgres
	docker compose -f dev-compose.yml up --exit-code-from tests-postgres tests-postgres
	docker compose -f dev-compose.yml down

benchmarks:
	@echo "Launching Tests in Docker Compose"
	docker compose -f dev-compose.yml up -d cassandra-primary cassandra mysql
	sleep 120
	docker compose -f dev-compose.yml up --build benchmarks

clean:
	@echo "Cleaning up build junk"
	-docker compose -f dev-compose.yml down

tarmac:
	@echo "Starting Application"
	docker compose -f dev-compose.yml up --build tarmac

tarmac-performance: build
	@echo "Starting Application"
	docker compose -f dev-compose.yml up -d tarmac-performance
