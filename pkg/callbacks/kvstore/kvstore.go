/*
Package kvstore is part of the Tarmac suite of Host Callback packages. This package provides users with the ability to
provide WASM functions with a host callback interface that provides key:value storage capabilities.

	import (
		"github.com/tarmac-project/tarmac/pkg/callbacks"
		"github.com/tarmac-project/tarmac/pkg/callbacks/kvstore"
	)

	func main() {
		// Create instance of kvstore to register for callback execution
		kvStore := kvstore.New(kvstore.Config{})

		// Create Callback router and register httpclient
		router := callbacks.New()
		router.RegisterCallback("kvstore", "get", kvStore.Get)
		router.RegisterCallback("kvstore", "set", kvStore.Set)
		router.RegisterCallback("kvstore", "delete", kvStore.Delete)
	}
*/
package kvstore

import (
	"encoding/base64"
	"fmt"
	"github.com/madflojo/hord"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tarmac-project/tarmac"
)

// KVStore provides access to Host Callbacks that interact with the key:value datastores within Tarmac. The callbacks
// within KVStore provided all of the logic and error handlings of accessing and interacting with a key:value
// database. Users will send the specified JSON request and receive an appropriate JSON response.
type KVStore struct {
	// KV is the user-provided Key:Value store instance using the github.com/madflojo/hord package. This package by
	// itself does not manage database connections but rather relies on the hord.Database interface. Users must
	// supply an initiated hord.Database to work with.
	kv hord.Database
}

// Config is provided to users to configure the Host Callback. All Tarmac Callbacks follow the same configuration
// format; each Config struct gives the specific Host Callback unique functionality.
type Config struct {
	// KV is the user-provided Key:Value store instance using the github.com/madflojo/hord package. This package by
	// itself does not manage database connections but rather relies on the hord.Database interface. Users must
	// supply an initiated hord.Database to work with.
	KV hord.Database
}

// New will create and return a new KVStore instance that users can register as a Tarmac Host Callback function. Users
// can provide any custom KVStore configurations using the configuration options supplied.
func New(cfg Config) (*KVStore, error) {
	k := &KVStore{}
	if cfg.KV == nil {
		return k, fmt.Errorf("KV Store cannot be nil")
	}
	k.kv = cfg.KV
	return k, nil
}

// Get will fetch the stored data using the key specified within the incoming JSON. Logging, error handling, and
// base64 encoding of data are all handled via this function. Note, this function expects the KVStoreGetRequest
// JSON type as input and will return a KVStoreGetResponse JSON.
func (k *KVStore) Get(b []byte) ([]byte, error) {
	// Start Response Message assuming everything is good
	r := tarmac.KVStoreGetResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	// Parse incoming Request
	var rq tarmac.KVStoreGet
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = "Error Parsing Input"
	}

	// Fetch data from KVStore if we do not have any other errors
	if r.Status.Code == 200 {
		data, err := k.kv.Get(rq.Key)
		if err != nil {
			r.Status.Code = 404
			r.Status.Status = fmt.Sprintf("Unable to fetch key %s - %s", rq.Key, err)
		}

		// Encode Fetched Data to store within JSON
		r.Data = base64.StdEncoding.EncodeToString(data)
	}

	// Marshal a resposne JSON to return to caller
	rsp, err := ffjson.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal kvstore:get response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}

// Set will store data within the key:value datastore using the key specified within the incoming JSON. Logging, error
// handling, and base64 decoding of data are all handled via this function. Note, this function expects the
// KVStoreSetRequest JSON type as input and will return a KVStoreSetResponse JSON.
func (k *KVStore) Set(b []byte) ([]byte, error) {
	// Start Response Message assuming everything is good
	r := tarmac.KVStoreSetResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	// Parse incoming Request
	var rq tarmac.KVStoreSet
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = "Error Parsing Input"
	}

	// Decode data to store
	data, err := base64.StdEncoding.DecodeString(rq.Data)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = fmt.Sprintf("Unable to decode data - %s", err)
	}

	// Store data in KVStore if we do not have any other errors
	if r.Status.Code == 200 {
		err = k.kv.Set(rq.Key, data)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to store data using key %s - %s", rq.Key, err)
		}
	}

	// Marshal a resposne JSON to return to caller
	rsp, err := ffjson.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal kvstore:set response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}

// Delete will remove the key and data stored within the key:value datastore using the key specified within the incoming
// JSON. Logging and error handling are all handled via this callback function. Note, this function expects the
// KVStoreDeleteRequest JSON type as input and will return a KVStoreDeleteResponse JSON.
func (k *KVStore) Delete(b []byte) ([]byte, error) {
	// Start Response Message assuming everything is good
	r := tarmac.KVStoreDeleteResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	// Parse incoming Request
	var rq tarmac.KVStoreDelete
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = "Error Parsing Input"
	}

	// Delete data in KVStore if we do not have any other errors
	if r.Status.Code == 200 {
		err = k.kv.Delete(rq.Key)
		if err != nil {
			r.Status.Code = 404
			r.Status.Status = fmt.Sprintf("Unable to delete key %s - %s", rq.Key, err)
		}
	}

	// Marshal a response JSON to return to caller
	rsp, err := ffjson.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal kvstore:delete response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}

// Keys will return a list of all keys stored within the key:value datastore. Logging and error handling are
// all handled via this callback function. Note, this function will return a KVStoreKeysResponse JSON.
func (k *KVStore) Keys([]byte) ([]byte, error) {
	// Start Response Message assuming everything is good
	r := tarmac.KVStoreKeysResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	// Fetch keys from datastore
	var err error
	r.Keys, err = k.kv.Keys()
	if err != nil {
		r.Status.Code = 500
		r.Status.Status = fmt.Sprintf("Unable to fetch keys - %s", err)
	}

	// Marshal a response JSON to return to caller
	rsp, err := ffjson.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal kvstore:delete response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}
