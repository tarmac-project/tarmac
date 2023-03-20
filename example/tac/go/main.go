// Tac is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload in reverse.
package main

import (
	"fmt"
	"github.com/madflojo/tarmac/pkg/sdk"
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
	var rsp []byte
	var err error

	// Log it
	tarmac.Logger.Trace(fmt.Sprintf("Reversing Payload: %s", payload))

	// Check Cache
	rsp, err = tarmac.KV.Get(string(payload))
	if err != nil || len(rsp) < 1 {
		// Flip it and reverse
		rsp = payload
		if len(rsp) > 0 {
			for i, n := 0, len(rsp)-1; i < n; i, n = i+1, n-1 {
				rsp[i], rsp[n] = rsp[n], rsp[i]
			}
		}

		// Store in Cache
		err = tarmac.KV.Set(string(payload), rsp)
		if err != nil {
			tarmac.Logger.Error(fmt.Sprintf("Unable to cache reversed payload: %s", err))
			return rsp, nil
		}
	}

	// Return the rsp
	return rsp, nil
}
