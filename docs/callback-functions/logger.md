---
description: Standard logging capabilities
---

# Logger

The Logger capability provides WASM functions with the ability to log arbitrary data. Much like a traditional logger framework, this capability offers multiple log levels.

By default Debug and Trace level logs are disabled. To enable these log levels to consult the configuration options for Tarmac.

## Error

Critical errors within the system.

```golang
_, err := wapc.HostCall("tarmac", "logger", "error", []byte("This is the data that should be logged"))
```
### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `logger` | `error` | Log Message | `nil` |

## Warn

Non-critical errors within the system.

```golang
_, err := wapc.HostCall("tarmac", "logger", "warn", []byte("This is the data that should be logged"))
```
### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `logger` | `warn` | Log Message | `nil` |

## Info

Informational logs.

```golang
_, err := wapc.HostCall("tarmac", "logger", "info", []byte("This is the data that should be logged"))
```
### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `logger` | `info` | Log Message | `nil` |

## Debug

Request level errors & informational logs. Disabled by default, calls to Debug logging are ignored unless `debug` is enabled for the Tarmac host.

```golang
_, err := wapc.HostCall("tarmac", "logger", "debug", []byte("This is the data that should be logged"))
```
### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `logger` | `debug` | Log Message | `nil` |

## Trace

Low-level details of execution. Disabled by default, calls to Trace logging are ignored unless `trace` is enabled for the Tarmac host.

```golang
_, err := wapc.HostCall("tarmac", "logger", "trace", []byte("This is the data that should be logged"))
```

### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `logger` | `trace` | Log Message | `nil` |
