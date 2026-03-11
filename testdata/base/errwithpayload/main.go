// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"fmt"

	"github.com/tarmac-project/tarmac/pkg/sdk"
)

func main() {
	// Initialize the Tarmac SDK
	_, err := sdk.New(sdk.Config{Namespace: "tarmac", Handler: Handler})
	if err != nil {
		return
	}
}

func Handler(_ []byte) ([]byte, error) {
	// Return a payload alongside an error to test that HTTP handler writes the body on error path
	return []byte("error with payload"), fmt.Errorf("this is a test error with payload")
}
