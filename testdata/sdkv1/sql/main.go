// This program is a test program used to facilitate unit testing with Tarmac.
package main

import (
	"fmt"

	"github.com/tarmac-project/tarmac/pkg/sdk"
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
	// SQL Query
	_, err := tarmac.SQL.Query(`CREATE TABLE IF NOT EXISTS wasmguest ( id int NOT NULL, name varchar(255), PRIMARY KEY (id) );`)
	if err != nil {
		tarmac.Logger.Error(fmt.Sprintf("Unable to execute SQL query - %s", err))
		return []byte(""), fmt.Errorf(`Failed to SQL callback - %s`, err)
	}

	// Return a happy message
	return []byte("Howdie"), nil
}
