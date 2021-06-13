/*
Package tarmac is a client package for WASM modules running within a Tarmac server.

This package provides user-friendly functions that wrap the Web Assembly Procedure Call (waPC) based functions of Tarmac. Guest WASM modules running inside Tarmac can use this library to call back the Tarmac host and perform host-level actions such as storing data within the database, logging specific data, or looking up configurations.
*/
package tarmac

// Request is used to create the Payload JSON sent to Tarmac WASM modules for incoming requests. A "request:handler" function will recieve this Request type
// in JSON format and will need to parse it accordingly.
type Request struct {

	// Headers are request headers such as HTTP headers, or other metadata depending on the protocol the request as received with.
	Headers map[string]string `json:"headers"`

	// Payload is a []byte that has been base64 encoded.
	Payload string `json:"payload"`
}

// Response is used to create a Response Payload sent from Tarmac WASM modules to the Tarmac host. A "request:handler" function will return this Response type in JSON
// format. The host will parse it accordingly and use the response details to reply to clients.
type Response struct {

	// Headers are response headers such as HTTP headers, or other metadata depending on the protocol. Any value in this map will be appended to existing response headers provided by
	// the Tarmac server.
	Headers map[string]string `json:"headers"`

	// StatusCode is the response code to return to the Tarmac host to indicate failure or success of the WASM call. The status code should follow HTTP status code conventions
	// even if the original request was made using non-HTTP protocols, the Tarmac server will translate the status code to a protocol appropriate return code.
	StatusCode int `json:"status_code"`

	// Payload is a []byte that should be base64 encoded. This payload is the response payload from the WASM Module itself. The contents of this field will be decoded and
	// returned to the client.
	Payload string `json:"payload"`
}
