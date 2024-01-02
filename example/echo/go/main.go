// Echo is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload.
package main

import (
	"fmt"

	"github.com/tarmac-project/tarmac/pkg/sdk"
)

// tarmac provides an interface for executing host capabilities such as Logging, KVStore, etc.
var tarmac *sdk.Tarmac

func main() {
	var err error

	// Initialize SDK
	tarmac, err = sdk.New(sdk.Config{Handler: Handler})
	if err != nil {
		return
	}
}

// Handler is the custom Tarmac Handler function that will receive a payload and
// must return a payload along with a nil error.
func Handler(payload []byte) ([]byte, error) {
	// Log It
	tarmac.Logger.Trace(fmt.Sprintf("Echoing Payload: %s", payload))

	// Return the payload
	return payload, nil
}
