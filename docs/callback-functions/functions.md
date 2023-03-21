---
description: Function to Function calls
---

# Functions

The Functions capability allows WASM functions to call other WASM functions. Unlike other callback capabilities, functions must register within the routes configuration within the `tarmac.json` configuration file.

## Function

```golang
_, err := wapc.HostCall("tarmac", "function", "function-name", []byte("Input to Function"))
```

### Interface Details

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `function` | Function Name | Input Data | Function Output Data |
