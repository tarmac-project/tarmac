// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"fmt"
	"strconv"

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
	// Fetch the counter value from KV datastore
	counter, err := tarmac.KV.Get("counter")
	if err != nil {
		// Assume that the counter does not exist
		tarmac.Logger.Info(`Counter does not exist, creating a new one`)
		err = tarmac.KV.Set("counter", []byte("0"))
		if err != nil {
			tarmac.Logger.Error(fmt.Sprintf("Failed to store counter via KVStore - %s", err))
			return []byte(""), fmt.Errorf(`Failed to store counter via KVStore - %s`, err)
		}
	}

	// Increment the counter
	counterInt, err := strconv.Atoi(string(counter))
	if err != nil {
		return []byte(""), fmt.Errorf(`Failed to convert counter to int - %s`, err)
	}
	counterInt++
	tarmac.Logger.Debug(fmt.Sprintf("Counter is now %d", counterInt))

	// Store the new counter value
	err = tarmac.KV.Set("counter", []byte(strconv.Itoa(counterInt)))
	if err != nil {
		tarmac.Logger.Error(fmt.Sprintf("Failed to store counter via KVStore - %s", err))
		return []byte(""), fmt.Errorf(`Failed to store counter via KVStore - %s`, err)
	}
	tarmac.Logger.Debug(fmt.Sprintf("Stored new counter value %d", counterInt))

	// If the counter is less than 5, return an error
	if counterInt < 5 {
		return []byte(""), fmt.Errorf(`Counter is less than 5`)
	}

	return []byte("Counter is greater than 5"), nil
}
