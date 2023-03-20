package kvstore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
)

type KVSetTestCase struct {
	name     string
	err      bool
	key      string
	data     []byte
	hostCall func(string, string, string, []byte) ([]byte, error)
}

func TestKVStore(t *testing.T) {
	tc := []KVSetTestCase{
		{
			name: "Key empty",
			err:  true,
			key:  "",
			data: []byte("Do i matter?"),
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				t.Errorf("HostCall should not be called")
				return []byte(""), nil
			},
		},
		{
			name: "Data empty",
			err:  true,
			key:  "test_key",
			data: []byte(""),
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				t.Errorf("HostCall should not be called")
				return []byte(""), nil
			},
		},
		{
			name: "Valid key and data",
			err:  false,
			key:  "test_key",
			data: []byte("test_data"),
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				return []byte(""), nil
			},
		},
		{
			name: "HostCall error",
			err:  true,
			key:  "foo",
			data: []byte("bar"),
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				return nil, fmt.Errorf("HostCall error")
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			kv, err := New(Config{Namespace: "default", HostCall: c.hostCall})
			if err != nil {
				t.Errorf("Unexpected error initializing kvstore - %s", err)
			}

			err = kv.Set(c.key, c.data)
			if err != nil && !c.err {
				t.Errorf("Unexpected error executing test case - %s", err)
				return
			}
			if c.err && err == nil {
				t.Errorf("Unexpected success executing kv.Set() that should error")
			}
		})
	}
}

type KVGetTestCase struct {
	name     string
	err      bool
	key      string
	data     []byte
	hostCall func(string, string, string, []byte) ([]byte, error)
}

func TestKVStore_Get(t *testing.T) {
	tc := []KVGetTestCase{
		{
			name: "Key empty",
			err:  true,
			key:  "",
			data: []byte("Do i matter?"),
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				t.Errorf("HostCall should not be called")
				return []byte(""), nil
			},
		},
		{
			name: "Key not found",
			err:  true,
			key:  "unknown_key",
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				return []byte(`{"status": {"code": 404, "status":"Unable to fetch key unknown_key - not found"}}`), fmt.Errorf("Unable to fetch key unknown_key - not found")
			},
		},
		{
			name: "Error executing hostcall",
			err:  true,
			key:  "some_key",
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				return nil, fmt.Errorf("some error")
			},
		},
		{
			name: "Valid key",
			err:  false,
			key:  "valid_key",
			data: []byte("hello"),
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				return []byte(fmt.Sprintf(`{"data":"%s"}`, base64.StdEncoding.EncodeToString([]byte("hello")))), nil
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			kv, err := New(Config{Namespace: "default", HostCall: c.hostCall})
			if err != nil {
				t.Errorf("Unexpected error initializing kvstore - %s", err)
			}

			data, err := kv.Get(c.key)
			if err != nil && !c.err {
				t.Errorf("Unexpected error executing test case - %s", err)
				return
			}
			if c.err && err == nil {
				t.Errorf("Unexpected success executing kv.Get() that should error")
			}
			if !c.err && !bytes.Equal(data, c.data) {
				t.Errorf("Expected data to be '%s' but got '%s'", string(c.data), string(data))
			}
		})
	}
}

func TestKVStore_Keys(t *testing.T) {
	hostCall := func(string, string, string, []byte) ([]byte, error) {
		return []byte(`{"keys":["key1", "key2"]}`), nil
	}
	kv, err := New(Config{Namespace: "default", HostCall: hostCall})
	if err != nil {
		t.Errorf("Unexpected error initializing kvstore - %s", err)
	}

	keys, err := kv.Keys()
	if err != nil {
		t.Errorf("Unexpected error executing kv.Keys() - %s", err)
		return
	}
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, but got %d", len(keys))
		return
	}
	if keys[0] != "key1" {
		t.Errorf("Expected first key to be 'key1', but got '%s'", keys[0])
	}
	if keys[1] != "key2" {
		t.Errorf("Expected second key to be 'key2', but got '%s'", keys[1])
	}
}

type KVDeleteTestCase struct {
	name     string
	err      bool
	key      string
	hostCall func(string, string, string, []byte) ([]byte, error)
}

func TestKVStore_Delete(t *testing.T) {
	tc := []KVDeleteTestCase{
		{
			name: "Key empty",
			err:  true,
			key:  "",
			hostCall: func(string, string, string, []byte) ([]byte, error) {
				t.Errorf("HostCall should not be called")
				return []byte(""), nil
			},
		},
		{
			name: "Delete success",
			err:  false,
			key:  "test",
			hostCall: func(namespace, operation, key string, data []byte) ([]byte, error) {
				if namespace != "default" || operation != "kvstore" || key != "delete" {
					t.Errorf("Incorrect arguments to hostCall")
				}
				if string(data) != `{"key":"test"}` {
					t.Errorf("Incorrect data passed to hostCall")
				}
				return []byte(`{"success": true}`), nil
			},
		},
		{
			name: "Delete failure",
			err:  true,
			key:  "test",
			hostCall: func(namespace, operation, key string, data []byte) ([]byte, error) {
				if namespace != "default" || operation != "kvstore" || key != "delete" {
					t.Errorf("Incorrect arguments to hostCall")
				}
				if string(data) != `{"key":"test"}` {
					t.Errorf("Incorrect data passed to hostCall")
				}
				return []byte(""), fmt.Errorf("hostCall failed")
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			kv, err := New(Config{Namespace: "default", HostCall: c.hostCall})
			if err != nil {
				t.Errorf("Unexpected error initializing kvstore - %s", err)
			}

			err = kv.Delete(c.key)
			if err != nil && !c.err {
				t.Errorf("Unexpected error executing test case - %s", err)
				return
			}
			if c.err && err == nil {
				t.Errorf("Unexpected success executing kv.Delete() that should error")
			}
		})
	}
}
