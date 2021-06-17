package app

import (
	"encoding/base64"
	"fmt"
	"github.com/madflojo/tarmac"
	"github.com/pquerna/ffjson/ffjson"
)

type kvStore struct{}

func (k *kvStore) Get(b []byte) ([]byte, error) {
	r := tarmac.KVStoreGetResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	var rq tarmac.KVStoreGet
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = "Error Parsing Input"
	}

	// Fetch data from KVStore
	data, err := kv.Get(rq.Key)
	if err != nil {
		r.Status.Code = 404
		r.Status.Status = fmt.Sprintf("Unable to fetch key %s - %s", rq.Key, err)
	}

	// Encode Fetched Data
	r.Data = base64.StdEncoding.EncodeToString(data)

	rsp, err := ffjson.Marshal(r)
	if err != nil {
		log.Errorf("Unable to marshal kvstore:get response - %s", err)
		return []byte(""), fmt.Errorf("unable to marshal kvstore:get response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}

func (k *kvStore) Set(b []byte) ([]byte, error) {
	r := tarmac.KVStoreSetResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

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

	// Store data in KVStore
	if r.Status.Code == 200 {
		err = kv.Set(rq.Key, data)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to store data using key %s - %s", rq.Key, err)
		}
	}

	rsp, err := ffjson.Marshal(r)
	if err != nil {
		log.Errorf("Unable to marshal kvstore:set response - %s", err)
		return []byte(""), fmt.Errorf("unable to marshal kvstore:set response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}

func (k *kvStore) Delete(b []byte) ([]byte, error) {
	r := tarmac.KVStoreDeleteResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	var rq tarmac.KVStoreDelete
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = "Error Parsing Input"
	}

	err = kv.Delete(rq.Key)
	if err != nil {
		r.Status.Code = 404
		r.Status.Status = fmt.Sprintf("Unable to delete key %s - %s", rq.Key, err)
	}

	rsp, err := ffjson.Marshal(r)
	if err != nil {
		log.Errorf("Unable to marshal kvstore:delete response - %s", err)
		return []byte(""), fmt.Errorf("unable to marshal kvstore:delete response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}
