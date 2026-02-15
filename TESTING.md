# Testing Guide

This document describes the testing structure and how to run tests in the Tarmac project.

## Test Types

Tarmac tests are organized into two categories:

### Unit Tests

Unit tests run locally without requiring external services like Redis, MySQL, or Consul. These tests are fast, deterministic, and suitable for local development.

**Running unit tests:**

```bash
make tests-unit
```

Or directly with go:

```bash
go test -v -race ./...
```

Unit tests include:
- Configuration validation tests
- Basic server functionality tests (health checks, metrics, pprof)
- Callback function tests with mocks
- WASM module loading tests

### Integration Tests

Integration tests require external services and are marked with the `integration` build tag. These tests verify that Tarmac works correctly with real databases, key-value stores, and configuration services.

**Running integration tests:**

```bash
make tests-integration
```

This runs all integration test suites with Docker Compose, including:
- `make tests-base` - Tests with Redis, MySQL, and Consul
- `make tests-redis` - Redis-specific tests
- `make tests-mysql` - MySQL-specific tests
- `make tests-postgres` - PostgreSQL-specific tests
- `make tests-cassandra` - Cassandra-specific tests
- `make tests-boltdb` - BoltDB-specific tests
- `make tests-inmemory` - In-memory KV store tests

**Running a specific integration test suite:**

```bash
make tests-redis    # Run only Redis tests
make tests-mysql    # Run only MySQL tests
# etc.
```

## Test Organization

- **Unit tests**: Located in `*_test.go` files without build tags
- **Integration tests**: Located in `*_integration_test.go` files with `//go:build integration` tag

Example integration test file:

```go
//go:build integration

package mypackage

import "testing"

func TestWithExternalService(t *testing.T) {
    // Test code that requires Redis, MySQL, etc.
}
```

## Best Practices

1. **Write unit tests first**: Most functionality should be testable without external services using mocks or fakes.

2. **Use integration tests sparingly**: Only use integration tests when you need to verify actual service integration.

3. **Fast feedback loop**: Run unit tests frequently during development (`make tests-unit`). Run integration tests before pushing changes.

4. **CI/CD**: The CI pipeline runs both unit and integration tests. Unit tests provide quick feedback, while integration tests ensure compatibility with real services.

## Timeouts and Retries

Integration tests use polling with retries instead of fixed sleeps to reduce flakiness:

```go
// Instead of:
time.Sleep(10 * time.Second)

// Use:
if err := waitForServer("http://localhost:9000/health", 15*time.Second); err != nil {
    t.Fatalf("Server failed to start: %v", err)
}
```

This approach:
- Reduces test runtime when services start quickly
- Provides more reliable tests by polling for readiness
- Fails faster when services don't start
