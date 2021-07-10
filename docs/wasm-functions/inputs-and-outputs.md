---
description: Server Request and Server Response Payloads
---

# Inputs & Outputs

When Tarmac executes a WASM function, it calls the function with a byte slice \(or language equivalent\) of raw JSON data called a Server Request. The WASM function must then return a Server Response JSON to Tarmac.

This Server Request and Server Response is how Tarmac can exchange request information back and forth with the running WASM function. Tarmac is geared to support many different request patterns, from HTTP APIs to Message Queues or even scheduled jobs. Regardless of how the WASM function is triggered, the Server Request and Server Response JSON payloads will remain consistent.

### Server Request

The below example JSON shows a Sever Request triggered from an HTTP POST request.

```javascript
{
	"headers": {
		"accept": "*/*",
		"content-length": "14",
		"content-type": "application/x-www-form-urlencoded",
		"http_method": "POST",
		"http_path": "/",
		"remote_addr": "172.21.0.1:56058",
		"request_type": "http",
		"user-agent": "curl/7.64.1"
	},
	"payload": "VGFybWFjIEV4YW1wbGU="
}
```

This JSON includes headers that are both from the HTTP request itself and are added by Tarmac. The `request_type` header, for example is added by Tarmac to inform the WASM function how the request was made.

{% hint style="info" %}
The HTTP payload sent from the client will be base64 encoded before it is included in the Server Request JSON. This is to avoid any formatting conflicts with the Server Request JSON payload itself.
{% endhint %}

### Server Response

When a WASM function executes, it must return a Server Response JSON to Tarmac. This response is used by Tarmac to determine how to respond to the request initiator.

```javascript
{
	"headers": {},
	"status": {
		"code": 200,
		"status": "Success"
	},
	"payload": "VGFybWFjIEV4YW1wbGU="
}
```

In the case of an HTTP call, any headers returned will be appended to the HTTP response, and the code from the status will be used as the HTTP return code. The HTTP status code is used even for non-HTTP-based requests. This allows WASM functions to have a consistent way of returning status without worrying about how the request originated.

The payload will be decoded and sent back to the HTTP requestor.

