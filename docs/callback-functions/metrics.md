---
description: Custom Metrics for WASM Functions
---

# Metrics

The Metrics capability provides WASM function developers the ability to create user-defined metrics exposed as part of the Tarmac `/metrics` end-point. This capability supports the three predominant metrics types, Counters, Gauges, & Histograms.

## Counter

The Counter function will give users the ability to create a custom counter metric. When called, this function will increment the counter by one.

```golang
_, err := wapc.HostCall("tarmac", "metrics", "counter", MetricsCounterJSON)
```

### Interface Details

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `metrics` | `counter` | `MetricsCounter` | `nil` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### MetricsCounter

```json
{
	"name": "custom_metric_name"
}
```

## Gauge

The Gauge function will give users the ability to create a custom gauge metric. When called, this function will either increment or decrement the gauge by one.

```golang
_, err := wapc.HostCall("tarmac", "metrics", "gauge", MetricsGaugeJSON)
```

### Interface Details

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `metrics` | `gauge` | `MetricsGauge` | `nil` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### MetricsCounter

```json
{
	"name": "custom_metric_name",
	"action": "inc"
}
```

Valid actions are `inc` (Increment) and `dec` (Decrement).

## Histogram

The Histogram function will give users the ability to create a custom histogram metric. When called, this function will observe the provided value and summarize the metric results.


```golang
_, err := wapc.HostCall("tarmac", "metrics", "histogram", MetricsHistogramJSON)
```

### Interface Details

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `metrics` | `histogram` | `MetricsHistogram` | `nil` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### MetricsCounter

```json
{
	"name": "custom_metric_name",
	"value": 0.0001
}
```
