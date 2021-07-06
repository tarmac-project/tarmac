// Hello is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload.
package main

import (
	"fmt"
	"github.com/valyala/fastjson"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	// Tarmac uses waPC to facilitate WASM module execution. Modules must register their custom handlers under the
	// appropriate method as shown below.
	wapc.RegisterFunctions(wapc.Functions{
		// Register a GET request handler
		"http:GET": NoHandler,
		// Register a POST request handler
		"http:POST": Handler,
		// Register a PUT request handler
		"http:PUT": Handler,
		// Register a DELETE request handler
		"http:DELETE": NoHandler,
	})
}

// NoHandler is a custom Tarmac Handler function that will return a tarmac.ServerResponse JSON that denies
// the client request.
func NoHandler(payload []byte) ([]byte, error) {
	return []byte(`{"status":{"code":503,"status":"Not Implemented"}}`), nil
}

// Handler is the custom Tarmac Handler function that will receive a tarmac.ServerRequest JSON payload and
// must return a tarmac.ServerResponse JSON payload along with a nil error.
func Handler(payload []byte) ([]byte, error) {
	// Perform a host callback to log the incoming request
	_, err := wapc.HostCall("tarmac", "logger", "debug", payload)
	if err != nil {
		return []byte(fmt.Sprintf(`{"status":{"code":500,"status":"Failed to call host callback - %s"}}`, err)), nil
	}

	// Parse the JSON request
	rq, err := fastjson.ParseBytes(payload)
	if err != nil {
		return []byte(fmt.Sprintf(`{"status":{"code":500,"status":"Failed to call parse json - %s"}}`, err)), nil
	}

	// Return the payload via a ServerResponse JSON
	return []byte(fmt.Sprintf(`{"payload":"%s","status":{"code":200,"status":"Success"}}`, rq.GetStringBytes("payload"))), nil
}
