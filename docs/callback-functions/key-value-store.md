---
description: Store and Retrieve data from a Key:Value datastore
---

# Key:Value Store

The Key:Value Store capability provides WASM function developers the ability to store and retrieve data from Key:Value datastores. At the moment, Tarmac supports multiple Key:Value stores which can be enabled/disabled in the host configuration settings.

## Get

The Get function provides users with the ability to fetch data using the specified key. To avoid conflicts, the `data` key within the response, JSON, will be base64 encoded.

```golang
_, err := wapc.HostCall("tarmac", "kvstore", "get", KVStoreGetJSON)
```
### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `kvstore` | `get` | `KVStoreGet` | `KVStoreGetResponse` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### KVStoreGet

```json
{
	"key": "myKey"
}
```

#### KVStoreGetResponse

```json
{
	"data": "VHdlZXQgYWJvdXQgVGFybWFjIGlmIHlvdSB0aGluayBpdCdzIGF3ZXNvbWUu",
	"status": {
		"code": 200,
		"status": "OK"
	}
}
```

The Status structure within the response JSON denotes the success of the database call. The status code value follows the HTTP status code standards, with anything higher than 399 is an error.

## Set

The Set function provides users with the ability to store data within the Key:Value datastore. The `data` key within the request JSON must be base64 encoded.

```golang
_, err := wapc.HostCall("tarmac", "kvstore", "set", KVStoreSetJSON)
```
### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `kvstore` | `set` | `KVStoreSet` | `KVStoreSetResponse` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### KVStoreSet

```json
{
  "data": "VHdlZXQgYWJvdXQgVGFybWFjIGlmIHlvdSB0aGluayBpdCdzIGF3ZXNvbWUu",
  "key": "myKey"
}
```

#### KVStoreSetResponse

```json
{
  "status": {
    "code": 200,
    "status": "OK"
  }
}
```

The Status structure within the response JSON denotes the success of the database call. The status code value follows the HTTP status code standards, with anything higher than 399 is an error.

## Delete

The Delete function provides users with the ability to delete data stored within the Key:Value datastore.

```golang
_, err := wapc.HostCall("tarmac", "kvstore", "delete", KVStoreDeleteJSON)
```
### Interface Details:

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `kvstore` | `delete` | `KVStoreDelete` | `KVStoreDeleteResponse` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### KVStoreDelete

```json
{
  "key": "myKey"
}
```

#### KVStoreDeleteResponse

```json
{
  "status": {
    "code": 200,
    "status": "OK"
  }
}
```

The Status structure within the response JSON denotes the success of the database call. The status code value follows the HTTP status code standards, with anything higher than 399 is an error.

