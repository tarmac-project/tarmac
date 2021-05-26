# Contributing to Hord

Thank you for your interest in helping develop Hord. The time, skills, and perspectives you contribute to this project are valued.

# How can I contribute?

Bugs, Design Proposals, Feature Requests, and Questions are all welcome and can be submitted by creating a [Github Issue](https://github.com/madflojo/hord/issues/new/choose) using one of the templates provided. Please provide as much detail as you can.

Code contributions are welcome as well! In an effort to keep this project tidy, please:
- Use `dep` to install and lock dependencies
- Use `gofmt` to format code and tests
- Run `go vet -v ./...` to check for any inadvertent suspicious code
- Write and run unit tests when they make sense using `go test`
- Write and run integration tests where applicable using docker compose
	- Start Hord with Cassandra by running `docker-compose -f dev-compose.yml up -d cassandra-primary cassandra`
	- Run integration tests by running `docker-compose -f dev-compose.yml up --exit-code-from tests --build tests`
