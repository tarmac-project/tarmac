// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"github.com/tarmac-project/tarmac/pkg/sdk"
)

var tarmac *sdk.Tarmac

func main() {
	var err error

	// Initialize the Tarmac SDK
	tarmac, err = sdk.New(sdk.Config{Namespace: "tarmac", Handler: Handler})
	if err != nil {
		return
	}
}

func Handler(payload []byte) ([]byte, error) {
	// Call another function
	tarmac.Function.Call("logger", payload)

	// Return a happy message
	return payload, nil
}
