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

**Unit test runtime:** 
- `TestBadConfigs`: ~0.01s (configuration validation)
- `TestPProfServerEnabled`: ~31s (includes 30s profile test)
- `TestPProfServerDisabled`: ~0.5s (quick verification)
- **Total**: ~32s (down from 60s+ with fixed sleeps in integration tests)

Unit tests include:
- Configuration validation tests (`TestBadConfigs`)
- Server functionality tests (`TestPProfServerEnabled`, `TestPProfServerDisabled`)
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

**Running integration tests locally (requires services):**

```bash
# Requires Redis, MySQL, Consul running locally
go test -v -race -tags integration ./pkg/app
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

## pkg/app Test Structure

The `pkg/app` package has been refactored to separate unit and integration tests:

### Unit Tests (`app_test.go`)
- `TestBadConfigs` - Validates server fails correctly with invalid configurations
- `TestPProfServerEnabled` - Validates pprof endpoints when enabled
- `TestPProfServerDisabled` - Validates pprof endpoints are blocked when disabled

### Integration Tests (`app_integration_test.go`)
- `TestRunningServer` - Tests server with Redis integration
- `TestRunningTLSServer` - Tests TLS server with Redis, MySQL, and Consul
- `TestRunningMTLSServer` - Tests mTLS server with services
- `TestRunningFailMTLSServer` - Tests mTLS authentication failures

## Best Practices

1. **Write unit tests first**: Most functionality should be testable without external services using mocks or fakes.

2. **Use integration tests sparingly**: Only use integration tests when you need to verify actual service integration.

3. **Fast feedback loop**: Run unit tests frequently during development (`make tests-unit`). Run integration tests before pushing changes.

4. **CI/CD**: The CI pipeline runs both unit and integration tests. Unit tests provide quick feedback, while integration tests ensure compatibility with real services.

## Improvements Over Previous Implementation

### Fixed Sleeps Replaced with Polling

Integration tests now use polling with retries instead of fixed sleeps to reduce flakiness:

```go
// Instead of:
time.Sleep(10 * time.Second)

// Use:
if err := waitForServer("http://localhost:9000/health", 15*time.Second); err != nil {
    t.Fatalf("Server failed to start: %v", err)
}
```

Benefits:
- Reduces test runtime when services start quickly
- Provides more reliable tests by polling for readiness
- Fails faster when services don't start properly
- Typical wait time reduced from 10-15s fixed to ~0.5-5s actual

### Test Isolation

- Unit tests no longer require Docker Compose or external services
- Integration tests are explicitly marked and skipped by default
- Clear separation of concerns makes test failures easier to diagnose

### Path Handling

- Unit tests use relative paths (`../../testdata/`) to work locally
- Integration tests use absolute paths (`/testdata/`) for Docker environment
- Docker Compose mounts testdata to `/testdata` for consistency
