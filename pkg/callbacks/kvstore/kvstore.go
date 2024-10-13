/*
Package kvstore is part of the Tarmac suite of Host Callback packages. This package provides users with the ability to
provide WASM functions with a host callback interface that provides key:value storage capabilities.

	import (
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
	"errors"
	"fmt"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/tarmac-project/hord"
	"github.com/tarmac-project/tarmac"

	"github.com/tarmac-project/protobuf-go/sdk"
	proto "github.com/tarmac-project/protobuf-go/sdk/kvstore"
	pb "google.golang.org/protobuf/proto"
)

// KVStore provides access to Host Callbacks that interact with the key:value datastores within Tarmac. The callbacks
// within KVStore provided all of the logic and error handlings of accessing and interacting with a key:value
// database. Users will send the specified JSON request and receive an appropriate JSON response.
type KVStore struct {
	// KV is the user-provided Key:Value store instance using the github.com/tarmac-project/hord package. This package by
	// itself does not manage database connections but rather relies on the hord.Database interface. Users must
	// supply an initiated hord.Database to work with.
	kv hord.Database
}

// Config is provided to users to configure the Host Callback. All Tarmac Callbacks follow the same configuration
// format; each Config struct gives the specific Host Callback unique functionality.
type Config struct {
	// KV is the user-provided Key:Value store instance using the github.com/tarmac-project/hord package. This package by
	// itself does not manage database connections but rather relies on the hord.Database interface. Users must
	// supply an initiated hord.Database to work with.
	KV hord.Database
}

var (
	// ErrNilKey is returned when the key is nil
	ErrNilKey = errors.New("key cannot be nil")
)

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

// Get will fetch data from the key:value datastore using the key specified.
func (k *KVStore) Get(b []byte) ([]byte, error) {
	msg := &proto.KVStoreGet{}
	err := pb.Unmarshal(b, msg)
	if err != nil {
		return k.getJSON(b)
	}

	rsp := &proto.KVStoreGetResponse{
		Status: &sdk.Status{
			Code:   200,
			Status: "OK",
		},
	}

	if msg.Key == "" {
		rsp.Status.Code = 400
		rsp.Status.Status = ErrNilKey.Error()
	}

	if rsp.Status.Code == 200 {
		data, err := k.kv.Get(msg.Key)
		if err != nil {
			rsp.Status.Code = 404
			rsp.Status.Status = fmt.Sprintf("Unable to fetch key %s", err)
		}

		rsp.Data = data
	}

	m, err := pb.Marshal(rsp)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal kvstore:get response")
	}

	if rsp.Status.Code == 200 {
		return m, nil
	}

	return m, fmt.Errorf("%s", rsp.Status.Status)
}

// getJSON retains the JSON based Get function for backwards compatibility. This function will be removed in future
// versions of Tarmac.
func (k *KVStore) getJSON(b []byte) ([]byte, error) {
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

// Set will store data within the key:value datastore using the key specified.
func (k *KVStore) Set(b []byte) ([]byte, error) {
	msg := &proto.KVStoreSet{}
	err := pb.Unmarshal(b, msg)
	if err != nil {
		return k.setJSON(b)
	}

	rsp := &proto.KVStoreSetResponse{
		Status: &sdk.Status{
			Code:   200,
			Status: "OK",
		},
	}

	if msg.Key == "" {
		rsp.Status.Code = 400
		rsp.Status.Status = ErrNilKey.Error()
	}

	if msg.Data == nil {
		rsp.Status.Code = 400
		rsp.Status.Status = "Data cannot be nil"
	}

	if rsp.Status.Code == 200 {
		err = k.kv.Set(msg.Key, msg.Data)
		if err != nil {
			rsp.Status.Code = 500
			rsp.Status.Status = fmt.Sprintf("Unable to store data - %s", err)
		}
	}

	m, err := pb.Marshal(rsp)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal kvstore:set response")
	}

	if rsp.Status.Code == 200 {
		return m, nil
	}

	return m, fmt.Errorf("%s", rsp.Status.Status)
}

// setJSON retains the JSON based Set function for backwards compatibility. This function will be removed in future
// versions of Tarmac.
func (k *KVStore) setJSON(b []byte) ([]byte, error) {
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

// Delete will remove data from the key:value datastore using the key specified.
func (k *KVStore) Delete(b []byte) ([]byte, error) {
	msg := &proto.KVStoreDelete{}
	err := pb.Unmarshal(b, msg)
	if err != nil {
		return k.deleteJSON(b)
	}

	rsp := &proto.KVStoreDeleteResponse{
		Status: &sdk.Status{
			Code:   200,
			Status: "OK",
		},
	}

	if msg.Key == "" {
		rsp.Status.Code = 400
		rsp.Status.Status = ErrNilKey.Error()
	}

	if rsp.Status.Code == 200 {
		err = k.kv.Delete(msg.Key)
		if err != nil {
			rsp.Status.Code = 404
			rsp.Status.Status = fmt.Sprintf("Unable to delete key %s", err)
		}
	}

	m, err := pb.Marshal(rsp)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal kvstore:delete response")
	}

	if rsp.Status.Code == 200 {
		return m, nil
	}

	return m, fmt.Errorf("%s", rsp.Status.Status)
}

// deleteJSON retains the JSON based Delete function for backwards compatibility. This function will be removed in future
// versions of Tarmac.
func (k *KVStore) deleteJSON(b []byte) ([]byte, error) {
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

func (k *KVStore) Keys(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return k.keysJSON(b)
	}

	rsp := &proto.KVStoreKeysResponse{
		Status: &sdk.Status{
			Code:   200,
			Status: "OK",
		},
	}

	keys, err := k.kv.Keys()
	if err != nil {
		rsp.Status.Code = 500
		rsp.Status.Status = fmt.Sprintf("Unable to fetch keys - %s", err)
	}
	rsp.Keys = keys

	m, err := pb.Marshal(rsp)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal kvstore:keys response")
	}

	if rsp.Status.Code == 200 {
		return m, nil
	}

	return m, fmt.Errorf("%s", rsp.Status.Status)
}

func (k *KVStore) keysJSON(_ []byte) ([]byte, error) {
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
