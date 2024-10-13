// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"fmt"

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
	// Store data within KV datastore
	err := tarmac.KV.Set("test-data", []byte("i am a little teapot"))
	if err != nil {
		return []byte(""), fmt.Errorf(`Failed to store data via KVStore - %s`, err)
	}

	// Fetch data from KV datastore
	data, err := tarmac.KV.Get("test-data")
	if err != nil {
		return []byte(""), fmt.Errorf(`Failed to fetch data via KVStore - %s`, err)
	}

	tarmac.Logger.Info(fmt.Sprintf("Fetched %s from datastore", data))

	if len(data) != len([]byte("i am a little teapot")) {
		return []byte(""), fmt.Errorf("not able to fetch data from KVStore")
	}

	// Return a happy message
	return []byte("Howdie"), nil
}
