// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"fmt"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	// Tarmac uses waPC to facilitate WASM module execution. Modules must register their custom handlers
	wapc.RegisterFunctions(wapc.Functions{
		"handler": Handler,
	})
}

func Handler(payload []byte) ([]byte, error) {
	// Log the payload
	_, err := wapc.HostCall("tarmac", "logger", "info", []byte(`Testdata Function Starting Execution`))
	if err != nil {
		return []byte(""), fmt.Errorf("Unable to call host callback - %s", err)
	}

	// Return a happy message
	return []byte("Howdie"), nil
}
