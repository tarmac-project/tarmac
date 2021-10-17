/*
This example WASM module shows how Users can add tarmac Callback capabilities to any Go host.
*/
package main

import (
	"fmt"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	// Register the Example function for execution. Multiple functions can be registered with different call names.
	wapc.RegisterFunctions(wapc.Functions{
		"example": Example,
	})
}

// Example is a simple function that adheres to the wapc signature.
func Example(payload []byte) ([]byte, error) {
	// Execute Host Callback to log
	_, err := wapc.HostCall("tarmac", "logger", "info", payload)
	if err != nil {
		return []byte(""), fmt.Errorf("Failure - %s", err)
	}
	return []byte("Success"), nil
}
