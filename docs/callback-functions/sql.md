---
description: Store and Retrieve data from a SQL datastore
---

# SQL Datastore

The SQL Datastore capability provides WASM function developers the ability to store and retrieve data from SQL datastores. At the moment, Tarmac supports multiple SQL stores which can be enabled/disabled in the host configuration settings.

## Query

The Query function provides users with the ability to execute custom SQL queries against the database service. The returned data is in JSON format and base64 encoded to avoid format conflicts.

```golang
_, err := wapc.HostCall("tarmac", "sql", "query", `SQLQueryJSON`)
```
### Interface Details

| Namespace | Capability | Function | Input | Output |
| --------- | ---------- | -------- | ----- | ------ |
| `tarmac` | `sql` | `query` | `SQLQuery` | `SQLQueryResponse` |

### Example JSON

This callback uses JSON messages as input and output to facilitate communications between WASM functions and the Tarmac host.

#### SQLQuery

To avoid format and data conflicts the query itself must be base64 encoded.

```json
{
	"query": "c2VsZWN0ICogZnJvbSBleGFtcGxlOw=="
}
```

#### SQLQueryResponse

To avoid format and data conflicts the data returned is base64 encoded.

```json
{
	"data": "W3siaWQiOjEsIm5hbWUiOiJKb2huIFNtaXRoIn0seyJpZCI6MSwibmFtZSI6IkphbmUgU21pdGgifV0=",
	"status": {
		"code": 200,
		"status": "OK"
	}
}
```

The Status structure within the response JSON denotes the success of the database call. The status code value follows the HTTP status code standards, with anything higher than 399 is an error.
