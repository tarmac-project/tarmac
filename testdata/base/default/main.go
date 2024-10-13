// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"github.com/tarmac-project/tarmac/pkg/sdk"
)

func main() {
	// Initialize the Tarmac SDK
	_, err := sdk.New(sdk.Config{Namespace: "tarmac", Handler: Handler})
	if err != nil {
		return
	}
}

func Handler(payload []byte) ([]byte, error) {
	// Return a happy message
	return []byte("Howdie"), nil
}
