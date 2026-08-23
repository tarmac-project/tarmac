// This program verifies Tarmac's raw waPC KV and logging host call contracts.
package main

import (
	"bytes"
	"fmt"

	kvstore "github.com/tarmac-project/protobuf-go/sdk/kvstore"
	"github.com/wapc/wapc-guest-tinygo"
)

const (
	namespace = "tarmac"
	testKey   = "hostcall-contract"
	testValue = "kvlogging-ok"
)

func main() {
	registerHandler()
}

// Initialize registers the waPC entrypoint for reactor-style WASI modules.
//
//go:wasmexport wapc_init
func Initialize() {
	registerHandler()
}

func registerHandler() {
	wapc.RegisterFunction("handler", handler)
}

func handler([]byte) ([]byte, error) {
	setRequest, err := (&kvstore.KVStoreSet{
		Key:  testKey,
		Data: []byte(testValue),
	}).MarshalVT()
	if err != nil {
		return nil, fmt.Errorf("marshal kv set request: %w", err)
	}

	setPayload, callErr := wapc.HostCall(namespace, "kvstore", "set", setRequest)
	if err := validateSetResponse(setPayload, callErr); err != nil {
		return nil, err
	}

	getRequest, err := (&kvstore.KVStoreGet{Key: testKey}).MarshalVT()
	if err != nil {
		return nil, fmt.Errorf("marshal kv get request: %w", err)
	}

	getPayload, callErr := wapc.HostCall(namespace, "kvstore", "get", getRequest)
	data, err := validateGetResponse(getPayload, callErr)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(data, []byte(testValue)) {
		return nil, fmt.Errorf("kv value = %q, want %q", data, testValue)
	}

	if _, err := wapc.HostCall(namespace, "logger", "info", []byte("raw hostcall contracts verified")); err != nil {
		return nil, fmt.Errorf("call logger/info: %w", err)
	}

	return data, nil
}

func validateSetResponse(payload []byte, callErr error) error {
	var response kvstore.KVStoreSetResponse
	if err := response.UnmarshalVT(payload); err != nil {
		return fmt.Errorf("unmarshal kv set response: %w", err)
	}
	if status := response.GetStatus(); status == nil || status.GetCode() != 200 {
		if callErr != nil {
			return fmt.Errorf("kv set status = %v: %w", status, callErr)
		}
		return fmt.Errorf("kv set status = %v", status)
	}
	if callErr != nil {
		return fmt.Errorf("call kvstore/set: %w", callErr)
	}
	return nil
}

func validateGetResponse(payload []byte, callErr error) ([]byte, error) {
	var response kvstore.KVStoreGetResponse
	if err := response.UnmarshalVT(payload); err != nil {
		return nil, fmt.Errorf("unmarshal kv get response: %w", err)
	}
	if status := response.GetStatus(); status == nil || status.GetCode() != 200 {
		if callErr != nil {
			return nil, fmt.Errorf("kv get status = %v: %w", status, callErr)
		}
		return nil, fmt.Errorf("kv get status = %v", status)
	}
	if callErr != nil {
		return nil, fmt.Errorf("call kvstore/get: %w", callErr)
	}
	return response.GetData(), nil
}
