// Tac is a small, simple Go program that is an example WASM module for Tarmac. This program will accept a Tarmac
// server request, log it, and echo back the payload in reverse.
package main

import (
	"errors"
	"fmt"

	"github.com/tarmac-project/sdk"
	"github.com/tarmac-project/sdk/kv"
	"github.com/tarmac-project/sdk/logging"
)

var (
	logger  logging.Client
	kvStore kv.Client
)

func main() {
	// Initialize the Tarmac SDK
	runtime, err := sdk.New(sdk.Config{Handler: Handler})
	if err != nil {
		return
	}

	cfg := runtime.Config()
	logger, err = logging.New(logging.Config{SDKConfig: cfg})
	if err != nil {
		return
	}

	kvStore, err = kv.New(kv.Config{SDKConfig: cfg})
	if err != nil {
		return
	}
}

// Handler is the custom Tarmac Handler function that will receive a payload and
// must return a payload along with a nil error.
func Handler(payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return nil, errors.New("payload cannot be empty")
	}

	// Log it
	logger.Trace(fmt.Sprintf("Reversing Payload: %s", payload))

	// Check Cache
	key := string(payload)
	rsp, err := kvStore.Get(key)
	if err == nil {
		return rsp, nil
	}
	if !errors.Is(err, kv.ErrKeyNotFound) {
		return nil, fmt.Errorf("unable to read cached payload: %w", err)
	}

	// Flip it and reverse
	for i, n := 0, len(payload)-1; i < n; i, n = i+1, n-1 {
		payload[i], payload[n] = payload[n], payload[i]
	}
	rsp = payload

	// Store in Cache
	if err := kvStore.Set(key, payload); err != nil {
		logger.Error(fmt.Sprintf("Unable to cache reversed payload: %s", err))
	}

	// Return the payload
	return rsp, nil
}
