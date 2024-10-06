package kvstore

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/tarmac-project/hord/drivers/mock"
	"github.com/tarmac-project/tarmac"

	proto "github.com/tarmac-project/protobuf-go/sdk/kvstore"
	pb "google.golang.org/protobuf/proto"
)

type KVStoreCase struct {
	err     bool
	pass    bool
	name    string
	key     string
	value   []byte
	mockCfg mock.Config
}

func TestKVStore(t *testing.T) {
	td := map[string][]byte{
		"Happy Path Value":             []byte("somedata"),
		"Integer Value":                {0x00, 0x00, 0x00, 0x64},
		"String Value":                 []byte("Hello, World!"),
		"Float Value":                  {0x40, 0x49, 0x0f, 0xdb},
		"Boolean Value":                {0x01},
		"JSON Object":                  []byte(`{"name":"Test","msg":"Hi!"}`),
		"Binary Data":                  {0x00, 0x01, 0x02, 0x03, 0x04},
		"Special Characters in String": []byte("!@#$%^&*()_+"),
	}

	t.Run("Get", func(t *testing.T) {
		// Predefined test cases
		kc := []KVStoreCase{
			{
				err:   true,
				pass:  false,
				name:  "Key not found",
				key:   "",
				value: []byte(""),
				mockCfg: mock.Config{
					GetFunc: func(_ string) ([]byte, error) {
						return nil, fmt.Errorf("Key not found")
					},
				},
			},
		}

		// Dynamically generate test cases from the td map
		for k, v := range td {
			kc = append(kc, KVStoreCase{
				err:   false,
				pass:  true,
				name:  k,
				key:   k,
				value: v,
				mockCfg: mock.Config{
					GetFunc: func(key string) ([]byte, error) {
						if key == k {
							return v, nil
						}
						return nil, fmt.Errorf("Key not found")
					},
				},
			})
		}

		// Execute each test case
		for _, c := range kc {
			t.Run(c.name, func(t *testing.T) {
				kv, _ := mock.Dial(c.mockCfg)
				store, err := New(Config{KV: kv})
				if err != nil {
					t.Fatalf("Failed to create KVStore instance: %v", err)
				}

				req := &proto.KVStoreGet{Key: c.key}
				b, err := pb.Marshal(req)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}

				rsp, err := store.Get(b)
				if (err != nil) != c.err {
					t.Fatalf("Unexpected error state: %v", err)
				}

				var response proto.KVStoreGetResponse
				if err := pb.Unmarshal(rsp, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if (response.Status.Code == 200) != c.pass {
					t.Fatalf("Unexpected response status: %d", response.Status.Code)
				}

				if c.pass && !bytes.Equal(response.Data, c.value) {
					t.Fatalf("Unexpected response data: %v", response.Data)
				}
			})
		}
	})

	t.Run("Set", func(t *testing.T) {
		// Predefined test cases
		kc := []KVStoreCase{
			{
				err:   true,
				pass:  false,
				name:  "No Data",
				key:   "no-data",
				value: []byte(""),
				mockCfg: mock.Config{
					SetFunc: func(_ string, _ []byte) error {
						return nil
					},
				},
			},
			{
				err:   true,
				pass:  false,
				name:  "No Key",
				key:   "",
				value: []byte("some data"),
				mockCfg: mock.Config{
					SetFunc: func(_ string, _ []byte) error {
						return nil
					},
				},
			},
		}

		// Dynamically generate test cases from the td map
		for k, v := range td {
			kc = append(kc, KVStoreCase{
				err:   false,
				pass:  true,
				name:  k,
				key:   k,
				value: v,
				mockCfg: mock.Config{
					SetFunc: func(key string, data []byte) error {
						if key == k && bytes.Equal(data, v) {
							return nil
						}
						return fmt.Errorf("Error inserting data")
					},
				},
			})
		}

		// Execute each test case
		for _, c := range kc {
			t.Run(c.name, func(t *testing.T) {
				kv, _ := mock.Dial(c.mockCfg)
				k, err := New(Config{KV: kv})
				if err != nil {
					t.Fatalf("Failed to create KVStore instance: %v", err)
				}

				req := &proto.KVStoreSet{Key: c.key, Data: c.value}
				b, err := pb.Marshal(req)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}

				rsp, err := k.Set(b)
				if (err != nil) != c.err {
					t.Fatalf("Unexpected error state: %v", err)
				}

				var response proto.KVStoreSetResponse
				if err := pb.Unmarshal(rsp, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if (response.Status.Code == 200) != c.pass {
					t.Fatalf("Unexpected response status: %d", response.Status.Code)
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Predefined test cases
		kc := []KVStoreCase{
			{
				err:   true,
				pass:  false,
				name:  "Key not found",
				key:   "",
				value: []byte(""),
				mockCfg: mock.Config{
					DeleteFunc: func(_ string) error {
						return nil
					},
				},
			},
			{
				err:  false,
				pass: true,
				name: "Happy Path",
				key:  "testing-happy",
				mockCfg: mock.Config{
					DeleteFunc: func(key string) error {
						if key == "testing-happy" {
							return nil
						}
						return fmt.Errorf("Key not found")
					},
				},
			},
		}

		// Execute each test case
		for _, c := range kc {
			t.Run(c.name, func(t *testing.T) {
				kv, _ := mock.Dial(c.mockCfg)
				k, err := New(Config{KV: kv})
				if err != nil {
					t.Fatalf("Failed to create KVStore instance: %v", err)
				}

				req := &proto.KVStoreDelete{Key: c.key}
				b, err := pb.Marshal(req)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}

				rsp, err := k.Delete(b)
				if (err != nil) != c.err {
					t.Fatalf("Unexpected error state: %v", err)
				}

				var response proto.KVStoreDeleteResponse
				if err := pb.Unmarshal(rsp, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if (response.Status.Code == 200) != c.pass {
					t.Fatalf("Unexpected response status: %d", response.Status.Code)
				}
			})
		}
	})

	t.Run("Keys", func(t *testing.T) {
		// Predefined test cases
		kc := []KVStoreCase{
			{
				err:   true,
				pass:  false,
				name:  "Errored Keys",
				key:   "",
				value: []byte(""),
				mockCfg: mock.Config{
					KeysFunc: func() ([]string, error) {
						return []string{}, fmt.Errorf("Forced Error")
					},
				},
			},
			{
				err:   false,
				pass:  true,
				name:  "Happy Path",
				key:   "",
				value: []byte(""),
				mockCfg: mock.Config{
					KeysFunc: func() ([]string, error) {
						return []string{"key1", "key2", "key3"}, nil
					},
				},
			},
			{
				err:   false,
				pass:  true,
				name:  "Sooo Many Keys",
				key:   "",
				value: []byte(""),
				mockCfg: mock.Config{
					KeysFunc: func() ([]string, error) {
						keys := make([]string, 100)
						for i := 0; i < 100; i++ {
							keys[i] = fmt.Sprintf("key%d", i)
						}
						return keys, nil
					},
				},
			},
			{
				err:   false,
				pass:  true,
				name:  "No Keys",
				key:   "",
				value: []byte(""),
				mockCfg: mock.Config{
					KeysFunc: func() ([]string, error) {
						return []string{}, nil
					},
				},
			},
		}

		// Execute each test case
		for _, c := range kc {
			t.Run(c.name, func(t *testing.T) {
				kv, _ := mock.Dial(c.mockCfg)
				k, err := New(Config{KV: kv})
				if err != nil {
					t.Fatalf("Failed to create KVStore instance: %v", err)
				}

				req := &proto.KVStoreKeys{ReturnProto: true}
				b, err := pb.Marshal(req)
				if err != nil {
					t.Fatalf("Failed to marshal request: %v", err)
				}

				rsp, err := k.Keys(b)
				if (err != nil) != c.err {
					t.Fatalf("Unexpected error state: %v", err)
				}

				var response proto.KVStoreKeysResponse
				if err := pb.Unmarshal(rsp, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if (response.Status.Code == 200) != c.pass {
					t.Fatalf("Unexpected response status: %d", response.Status.Code)
				}
			})
		}
	})
}

type KVStoreCaseJSON struct {
	err  bool
	pass bool
	name string
	call string
	json string
}

func TestKVStoreJSON(t *testing.T) {
	// Set DB as a Mocked Database
	kv, _ := mock.Dial(mock.Config{
		GetFunc: func(key string) ([]byte, error) {
			if key == "testing-happy" {
				return []byte("somedata"), nil
			}
			return []byte(""), fmt.Errorf("Forced Error")
		},
		SetFunc: func(key string, _ []byte) error {
			if key == "testing-happy" {
				return nil
			}
			return fmt.Errorf("Error inserting data")
		},
		// Create a fake Delete function
		DeleteFunc: func(key string) error {
			if key == "testing-happy" {
				return nil
			}
			return fmt.Errorf("Error deleting data")
		},
		KeysFunc: func() ([]string, error) {
			return []string{}, fmt.Errorf("Forced Error")
		},
	})

	// Create new KVStore instance
	k, err := New(Config{KV: kv})
	if err != nil {
		t.Errorf("Unable to create new KVStore Instance - %s", err)
	}

	var kc []KVStoreCaseJSON

	// Create a collection of test cases
	kc = append(kc, KVStoreCaseJSON{
		err:  false,
		pass: true,
		name: "Happy Path",
		call: "Get",
		json: `{"key":"testing-happy"}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  false,
		pass: true,
		name: "Happy Path",
		call: "Set",
		json: `{"key":"testing-happy","data":"QmVjYXVzZSBJJ20gSGFwcHk="}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  false,
		pass: true,
		name: "Happy Path",
		call: "Delete",
		json: `{"key":"testing-happy"}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "Invalid Request JSON",
		call: "Get",
		json: `{"ke:"testing-happy"}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "Invalid Request JSON",
		call: "Set",
		json: `{"ke:"testing-happy","data":"QmVjYXVzZSBJJ20gSGFwcHk="}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "Invalid Request JSON",
		call: "Delete",
		json: `{"ke:"testing-happy"}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "Payload Not Base64",
		call: "Set",
		json: `{"key":"testing-happy","data":"not base64"}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "Key not found",
		call: "Get",
		json: `{"key":""}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "Failing Call",
		call: "Delete",
		json: `{"key": "invalid-key"}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "No Data",
		call: "Set",
		json: `{"key":"no-data"}`,
	})

	kc = append(kc, KVStoreCaseJSON{
		err:  true,
		pass: false,
		name: "Errored Keys",
		call: "Keys",
		json: ``,
	})

	// Loop through test cases executing and validating
	for _, c := range kc {
		switch c.call {
		case "Set":
			t.Run(c.name+" Set", func(t *testing.T) {
				// Set data first
				b, err := k.Set([]byte(c.json))
				if err != nil && !c.err {
					t.Fatalf("KVStore Callback Set failed unexpectedly - %s", err)
				}
				if err == nil && c.err {
					t.Fatalf("KVStore Callback Set unexpectedly passed")
				}

				// Validate Response
				var rsp tarmac.KVStoreSetResponse
				err = ffjson.Unmarshal(b, &rsp)
				if err != nil {
					t.Fatalf("KVStore Callback Set replied with an invalid JSON - %s", err)
				}

				if rsp.Status.Code == 200 && !c.pass {
					t.Fatalf("KVStore Callback Set returned an unexpected success - %+v", rsp)
				}
				if rsp.Status.Code != 200 && c.pass {
					t.Fatalf("KVStore Callback Set returned an unexpected failure - %+v", rsp)
				}
			})

		case "Get":
			t.Run(c.name+" Get", func(t *testing.T) {
				// Get data
				b, err := k.Get([]byte(c.json))
				if err != nil && !c.err {
					t.Fatalf("KVStore Callback Get failed unexpectedly - %s", err)
				}
				if err == nil && c.err {
					t.Fatalf("KVStore Callback Get unexpectedly passed")
				}

				// Validate Response
				var rsp tarmac.KVStoreGetResponse
				err = ffjson.Unmarshal(b, &rsp)
				if err != nil {
					t.Fatalf("KVStore Callback Get replied with an invalid JSON - %s", err)
				}

				if rsp.Status.Code == 200 && !c.pass {
					t.Fatalf("KVStore Callback Get returned an unexpected success - %+v", rsp)
				}
				if rsp.Status.Code != 200 && c.pass {
					t.Fatalf("KVStore Callback Get returned an unexpected failure - %+v", rsp)
				}
			})

		case "Delete":
			t.Run(c.name+" Delete", func(t *testing.T) {
				// Delete data
				b, err := k.Delete([]byte(c.json))
				if err != nil && !c.err {
					t.Fatalf("KVStore Callback Delete failed unexpectedly - %s", err)
				}
				if err == nil && c.err {
					t.Fatalf("KVStore Callback Delete unexpectedly passed")
				}

				// Validate Response
				var rsp tarmac.KVStoreDeleteResponse
				err = ffjson.Unmarshal(b, &rsp)
				if err != nil {
					t.Fatalf("KVStore Callback Delete replied with an invalid JSON - %s", err)
				}

				if rsp.Status.Code == 200 && !c.pass {
					t.Fatalf("KVStore Callback Delete returned an unexpected success - %+v", rsp)
				}
				if rsp.Status.Code != 200 && c.pass {
					t.Fatalf("KVStore Callback Delete returned an unexpected failure - %+v", rsp)
				}
			})

		case "Keys":
			t.Run(c.name+" Keys", func(t *testing.T) {
				// Fetch keys
				b, err := k.Keys([]byte(c.json))
				if err != nil && !c.err {
					t.Fatalf("KVStore Callback Keys failed unexpectedly - %s", err)
				}
				if err == nil && c.err {
					t.Fatalf("KVStore Callback Keys unexpectedly passed")
				}

				// Validate Response
				var rsp tarmac.KVStoreKeysResponse
				err = ffjson.Unmarshal(b, &rsp)
				if err != nil {
					t.Fatalf("KVStore Callback Keys replied with an invalid JSON - %s", err)
				}

				if rsp.Status.Code == 200 && !c.pass {
					t.Fatalf("KVStore Callback Keys returned an unexpected success - %+v", rsp)
				}
				if rsp.Status.Code != 200 && c.pass {
					t.Fatalf("KVStore Callback Keys returned an unexpected failure - %+v", rsp)
				}
			})
		}
	}
}
