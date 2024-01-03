// Tac is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload in reverse.
package main

import (
	"fmt"

	"github.com/tarmac-project/tarmac/pkg/sdk"
)

var tarmac *sdk.Tarmac

func main() {
	var err error

	// Initialize the Tarmac SDK
	tarmac, err = sdk.New(sdk.Config{Handler: Handler})
	if err != nil {
		return
	}
}

// Handler is the custom Tarmac Handler function that will receive a payload and
// must return a payload along with a nil error.
func Handler(payload []byte) ([]byte, error) {
	var err error

	// Log it
	tarmac.Logger.Trace(fmt.Sprintf("Reversing Payload: %s", payload))

	// Check Cache
	key := string(payload)
	rsp, err := tarmac.KV.Get(key)
	if err != nil || len(payload) < 1 {
		// Flip it and reverse
		if len(payload) > 0 {
			for i, n := 0, len(payload)-1; i < n; i, n = i+1, n-1 {
				payload[i], payload[n] = payload[n], payload[i]
			}
		}
		rsp = payload

		// Store in Cache
		err = tarmac.KV.Set(key, payload)
		if err != nil {
			tarmac.Logger.Error(fmt.Sprintf("Unable to cache reversed payload: %s", err))
			return rsp, nil
		}
	}

	// Return the payload
	return rsp, nil
}
