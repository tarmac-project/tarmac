# Makefile is used to drive the build and installation of this application
# this is meant to be used with a local copy of code repository.

tests:
	@echo "Launching Tests in Docker Compose"
	docker-compose -f dev-compose.yml up -d cassandra-primary cassandra
	sleep 30
	docker-compose -f dev-compose.yml up --build tests

benchmarks:
	@echo "Launching Tests in Docker Compose"
	docker-compose -f dev-compose.yml up -d cassandra-primary cassandra
	sleep 30
	docker-compose -f dev-compose.yml up --build benchmarks

clean:
	@echo "Cleaning up build junk"
	-docker-compose -f dev-compose.yml down

tarmac:
	@echo "Starting Application"
	docker-compose -f dev-compose.yml up --build tarmac
