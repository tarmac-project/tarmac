package app

import (
	"github.com/madflojo/hord/drivers/mock"
	"github.com/madflojo/tarmac"
	"github.com/pquerna/ffjson/ffjson"
	"testing"
)

type kvStoreCase struct {
	pass bool
	name string
	call string
	json string
}

func TestKVStore(t *testing.T) {
	k := &kvStore{}

	// Set DB as a Mocked Database
	kv, _ = mock.Dial(mock.Config{})

	var kc []kvStoreCase

	// Create a collection of test cases
	kc = append(kc, kvStoreCase{
		pass: true,
		name: "Happy Path",
		call: "Get",
		json: `{"key":"testing-happy"}`,
	})

	kc = append(kc, kvStoreCase{
		pass: true,
		name: "Happy Path",
		call: "Set",
		json: `{"key":"testing-happy","data":"QmVjYXVzZSBJJ20gSGFwcHk="}`,
	})

	kc = append(kc, kvStoreCase{
		pass: true,
		name: "Happy Path",
		call: "Delete",
		json: `{"key":"testing-happy"}`,
	})

	kc = append(kc, kvStoreCase{
		pass: false,
		name: "Invalid Request JSON",
		call: "Get",
		json: `{"ke:"testing-happy"}`,
	})

	kc = append(kc, kvStoreCase{
		pass: false,
		name: "Invalid Request JSON",
		call: "Set",
		json: `{"ke:"testing-happy","data":"QmVjYXVzZSBJJ20gSGFwcHk="}`,
	})

	kc = append(kc, kvStoreCase{
		pass: false,
		name: "Invalid Request JSON",
		call: "Delete",
		json: `{"ke:"testing-happy"}`,
	})

	kc = append(kc, kvStoreCase{
		pass: false,
		name: "Payload Not Base64",
		call: "Set",
		json: `{"key":"testing-happy","data":"not base64"}`,
	})

	// Loop through test cases executing and validating
	for _, c := range kc {
		switch c.call {
		case "Set":
			t.Run(c.name+" Set", func(t *testing.T) {
				// Set data first
				b, err := k.Set([]byte(c.json))
				if err != nil && c.pass {
					t.Fatalf("KVStore Callback Set failed unexpectedly - %s", err)
				}
				if err == nil && !c.pass {
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
				if err != nil && c.pass {
					t.Fatalf("KVStore Callback Get failed unexpectedly - %s", err)
				}
				if err == nil && !c.pass {
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
				if err != nil && c.pass {
					t.Fatalf("KVStore Callback Delete failed unexpectedly - %s", err)
				}
				if err == nil && !c.pass {
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

		}
	}
}
