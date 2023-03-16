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
	// SQL Query
	_, err := wapc.HostCall("tarmac", "sql", "query", []byte(`{"query":"Q1JFQVRFIFRBQkxFIElGIE5PVCBFWElTVFMgd2FzbWd1ZXN0ICggaWQgaW50IE5PVCBOVUxMLCBuYW1lIHZhcmNoYXIoMjU1KSwgUFJJTUFSWSBLRVkgKGlkKSApOw=="}`))
	if err != nil {
		return []byte(""), fmt.Errorf(`Failed to call host callback - %s`, err)
	}

	// Return a happy message
	return []byte("Howdie"), nil
}
