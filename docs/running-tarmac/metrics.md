---
description: Monitoring Tarmac with Prometheus Metrics
---

# Monitoring

Tarmac exposes several metrics to facilitate monitoring services. Metrics are available via the `/metrics` end-point in the Prometheus format.

These metrics include internal Tarmac system metrics such as the number of goroutines, memory utilization, and WASM function-specific metrics such as counters for Callbacks and WASM function execution time.

Some valuable metrics to monitor are in the below table.

| Metric Name | Metric Type | Description |
| ----------- | ----------- | ----------- |
| `http_server` | Summary | Summary of HTTP Server requests |
| `scheduled_tasks` | Summary | Summary of user defined scheduled task WASM function executions |
| `wasm_callbacks` | Summary | Summary of Tarmac callback function executions |
| `wasm_functions` | Summary | Summary of wasm function executions |

These metrics do not need to be enabled and are "on by default".
