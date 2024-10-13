// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"fmt"

	"github.com/tarmac-project/tarmac/pkg/sdk"
	wapc "github.com/wapc/wapc-guest-tinygo"

	"github.com/tarmac-project/protobuf-go/sdk/sql"
)

var tarmac *sdk.Tarmac

func main() {
	var err error

	// Initialize the Tarmac SDK
	tarmac, err = sdk.New(sdk.Config{Namespace: "tarmac", Handler: Handler})
	if err != nil {
		return
	}
}

func Handler(payload []byte) ([]byte, error) {
	// Create SQL Request
	query := &sql.SQLQuery{Query: []byte(`CREATE TABLE IF NOT EXISTS wasmguest ( id int NOT NULL, name varchar(255), PRIMARY KEY (id) );`)}
	q, err := query.MarshalVT()
	if err != nil {
		tarmac.Logger.Error(fmt.Sprintf("Unable to marshal SQL query - %s", err))
		return []byte(""), fmt.Errorf(`Failed to marshal SQL query - %s`, err)
	}

	// Call the SQL capability
	rsp, err := wapc.HostCall("tarmac", "sql", "query", q)
	if err != nil {
		tarmac.Logger.Error(fmt.Sprintf("Unable to call SQL capability - %s", err))
		return []byte(""), fmt.Errorf(`Failed to call SQL capability - %s`, err)
	}

	// Unmarshal the response
	var response sql.SQLQueryResponse
	err = response.UnmarshalVT(rsp)
	if err != nil {
		tarmac.Logger.Error(fmt.Sprintf("Unable to unmarshal SQL response - %s", err))
		return []byte(""), fmt.Errorf(`Failed to unmarshal SQL response - %s`, err)
	}

	// Validate the response
	if response.Status.Code != 200 {
		tarmac.Logger.Error(fmt.Sprintf("SQL query failed - %s", response.Status.Status))
		return []byte(""), fmt.Errorf(`SQL query failed - %s`, response.Status.Status)
	}

	// Return a happy message
	return []byte("Howdie"), nil
}
