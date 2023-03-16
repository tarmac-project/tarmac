// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	// Tarmac uses waPC to facilitate WASM module execution. Modules must register their custom handlers
	wapc.RegisterFunctions(wapc.Functions{
		"handler": Handler,
	})
}

func Handler(payload []byte) ([]byte, error) {
	// Return a happy message
	return []byte("Howdie"), nil
}
