# Makefile is used to drive the build and installation of this application
# this is meant to be used with a local copy of code repository.

tests:
	@echo "Launching Tests in Docker Compose"
	docker-compose -f dev-compose.yml up --build tests

clean:
	@echo "Cleaning up build junk"
	-docker-compose -f dev-compose.yml down

tarmac:
	@echo "Starting Application"
	docker-compose -f dev-compose.yml up --build tarmac
