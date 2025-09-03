---
description: Make HTTP requests with a simple HTTP client
---

# HTTPClient 

The HTTPClient capability provides WASM function developers to perform HTTP client requests to remote or local HTTP servers. While a simplistic client, Tarmac supports multiple HTTP requests, setting headers, and custom payloads.

## Security and Resource Management

To prevent DoS attacks and excessive memory usage, HTTP response bodies are limited to a configurable maximum size. By default, responses are limited to 10MB. This limit can be configured service-wide using the `http_client_max_response_body_size` configuration parameter (see [Configuration](../running-tarmac/configuration.md) for details). 

## Call

The Call function provides users with the ability to make HTTP client requests to the specified URL. The `body` key within the request and response JSON will be base64 encoded to avoid conflicts.

```golang
_, err := wapc.HostCall("tarmac", "httpclient", "call", HTTPClientJSON)
```
### Interface Details

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `httpclient` | `call` | `HTTPClient` | `HTTPClientResponse` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### HTTPClient

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

#### HTTPClientResponse

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

## Response Body Size Limiting

HTTP responses that exceed the configured maximum body size will be truncated to the limit. If the response body is larger than the configured limit, only the first bytes up to the limit will be included in the response. This behavior applies to both successful and error responses to ensure consistent resource management.
