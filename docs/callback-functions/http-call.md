---
description: Make HTTP requests with a simple HTTP client
---

# HTTP Call

The HTTP Call capability provides WASM function developers to perform HTTP client requests to remote or local HTTP servers. While a simplistic client, Tarmac supports multiple HTTP requests, setting headers, and custom payloads. 

## Call

The Call function provides users with the ability to make HTTP client requests to the specified URL. The `body` key within the request and response JSON will be base64 encoded to avoid conflicts.

```golang
_, err := wapc.HostCall("tarmac", "httpcall", "call", HTTPCallJSON)
```
### Interface Details

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `httpcall` | `call` | `HTTPCall` | `HTTPCallResponse` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### HTTPCall

```json
{
	"method": "POST",
	"headers": {
		"content-type": "application/json"
	},
	"insecure": true,
	"url": "http://example.com",
	"body": "ewoJIm1lIjogewoJCSJ0ZWFwb3QiOiB0cnVlCgl9Cn0="
}
```

#### HTTPCallResponse

```json
{
	"code": 400,
	"headers": {
		"key": "value"
	},
	"body": "dGlwIG1lIG92ZXIgYW5kIHBvdXIgbWUgb3V0",
	"status": {
		"code": 200,
		"status": "OK"
	}
}
```

The Status structure within the response JSON denotes the success of the database call. The status code value follows the HTTP status code standards, with anything higher than 399 is an error.

Note: Tarmac will provide a 200 status code if the HTTP request was successfully made, check the `code` value to validate the HTTP server return code.
