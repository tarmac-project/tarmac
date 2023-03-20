// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"fmt"
	"github.com/madflojo/tarmac/pkg/sdk"
)

var tarmac *sdk.Tarmac

func main() {
	var err error

	// Initialize the Tarmac SDK
	tarmac, err = sdk.New(sdk.Config{Namespace: "test-service", Handler: Handler})
	if err != nil {
		return
	}
}

func Handler(payload []byte) ([]byte, error) {
	// Store data within KV datastore
	err := tarmac.KV.Set("test-data", []byte("i am a little teapot"))
	if err != nil {
		return []byte(""), fmt.Errorf(`Failed to store data via KVStore - %s`, err)
	}

	// Fetch data from KV datastore
	_, err = tarmac.KV.Get("test-data")
	if err != nil {
		return []byte(""), fmt.Errorf(`Failed to fetch data via KVStore - %s`, err)
	}

	// Return a happy message
	return []byte("Howdie"), nil
}
