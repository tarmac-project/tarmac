# Contributing to Go Quick

Thank you for your interest in helping develop Go Quick. The time, skills, and perspectives you contribute to this project are valued.

## How can I contribute?

Bugs, Design Proposals, Feature Requests, and Questions are all welcome by creating a [Github Issue](https://github.com/madflojo/tarmac/issues/new/choose) using one of the templates provided. Please provide as much detail as you can.

Code contributions are welcome as well! To keep this project tidy, please:

- Use `go mod` to install and lock dependencies
- Use `gofmt` to format code and tests
- Run `go vet -v ./...` to check for any inadvertent suspicious code
- Write and run unit tests using `make test`

### Available Makefile commands

Run tests

```console
$ make test
```

Clean up environment

```console
$ make clean
```
