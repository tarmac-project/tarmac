/*
Package kvstore is a client package for WASM functions running within a Tarmac server.

This package provides a user-friendly Key:Value datastore interface that interacts with underlying datastores configured within Tarmac. Guest WASM functions running inside Tarmac can import and call this KV interface.
*/
package kvstore

import (
	"encoding/base64"
	"fmt"

	"github.com/valyala/fastjson"
)

// KV provides a simple interface for Tarmac Functions to store key:value data using supported KV datastores.
type KV struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// Config provides users with the ability to specify namespaces, function handlers and other key information required to execute the
// function.
type Config struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	Namespace string

	// HostCall is used internally for host callbacks. This is mainly here for testing.
	HostCall func(string, string, string, []byte) ([]byte, error)
}

// New creates a new KV with the provided configuration.
func New(cfg Config) (*KV, error) {
	// Set default namespace
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// Verify HostCall is set
	if cfg.HostCall == nil {
		return &KV{}, fmt.Errorf("HostCall cannot be nil")
	}

	return &KV{namespace: cfg.Namespace, hostCall: cfg.HostCall}, nil
}

// Set will store the supplied data under the provided key.
func (kv *KV) Set(key string, data []byte) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if len(data) == 0 {
		return fmt.Errorf("data cannot by empty")
	}

	d := base64.StdEncoding.EncodeToString(data)
	j := fmt.Sprintf(`{"key":"%s","data":"%s"}`, key, d)

	_, err := kv.hostCall(kv.namespace, "kvstore", "set", []byte(j))
	if err != nil {
		return fmt.Errorf("unable to execute set - %s", err)
	}

	return nil
}

// Get will fetch the data stored under the supplied key.
func (kv *KV) Get(key string) ([]byte, error) {
	if key == "" {
		return []byte(""), fmt.Errorf("key cannot be empty")
	}

	b, err := kv.hostCall(kv.namespace, "kvstore", "get", []byte(fmt.Sprintf(`{"key":"%s"}`, key)))
	if err != nil {
		return []byte(""), fmt.Errorf("unable to execute get - %s", err)
	}

	d, err := base64.StdEncoding.DecodeString(fastjson.GetString(b, "data"))
	if err != nil {
		return []byte(""), fmt.Errorf("unable to decode fetched data - %s", err)
	}

	return d, nil
}

// Delete will delete the data and key defined at key.
func (kv *KV) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	_, err := kv.hostCall(kv.namespace, "kvstore", "delete", []byte(fmt.Sprintf(`{"key":"%s"}`, key)))
	if err != nil {
		return fmt.Errorf("unable to execute delete - %s", err)
	}

	return nil
}

// Keys will return a list of keys available within the KV datastore.
func (kv *KV) Keys() ([]string, error) {
	var keys []string

	b, err := kv.hostCall(kv.namespace, "kvstore", "keys", []byte(""))
	if err != nil {
		return keys, fmt.Errorf("unable to execute keys - %s", err)
	}

	v, err := fastjson.ParseBytes(b)
	if err != nil {
		return keys, fmt.Errorf("unable to parse returned data - %s", err)
	}

	for _, k := range v.GetArray("keys") {
		keys = append(keys, string(k.GetStringBytes()))
	}

	return keys, nil
}
