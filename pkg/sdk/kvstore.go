package sdk

import (
	"encoding/base64"
	"fmt"
	"github.com/valyala/fastjson"
)

type KV struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

func newKVStore(cfg Config) *KV {
	return &KV{namespace: cfg.Namespace, hostCall: cfg.hostCall}
}

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
