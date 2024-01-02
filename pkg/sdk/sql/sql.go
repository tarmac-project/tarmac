/*
Package sql is a client package for WASM functions running within a Tarmac server.

This package provides a user-friendly SQL database interface with underlying databases configured within Tarmac. Guest WASM functions running inside Tarmac can import and call this SQL interface.
*/
package sql

import (
	"encoding/base64"
	"fmt"

	"github.com/valyala/fastjson"
)

// SQL provides an interface to the underlying SQL datastores within Tarmac.
type SQL struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// Config provides users with the ability to specify namespaces, function handlers and other key information required to execute the
// function.
type Config struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	Namespace string

	// HostCall is used internally for host callbacks. This is mainly here for testing.
	HostCall func(string, string, string, []byte) ([]byte, error)
}

// New returns a new instance of SQL initialized with Config.
func New(cfg Config) (*SQL, error) {
	// Set default namespace
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// Verify HostCall is set
	if cfg.HostCall == nil {
		return &SQL{}, fmt.Errorf("HostCall cannot be nil")
	}

	return &SQL{namespace: cfg.Namespace, hostCall: cfg.HostCall}, nil
}

// Query will execute the specified SQL query and return a byte array
// containing a JSON representation of the SQL data.
func (sql *SQL) Query(q string) ([]byte, error) {
	// Encode SQL Query
	qry := base64.StdEncoding.EncodeToString([]byte(q))

	// Callback to host
	rsp, err := sql.hostCall(sql.namespace, "sql", "query", []byte(fmt.Sprintf(`{"query":"%s"}`, qry)))
	if err != nil {
		return []byte(""), fmt.Errorf("error while executing host callback - %s", err)
	}

	// Fetch Data from JSON
	data := fastjson.GetString(rsp, "data")
	if data == "" {
		return []byte(""), nil
	}

	// Decode SQL Response
	d, err := base64.StdEncoding.DecodeString(fastjson.GetString(rsp, "data"))
	if err != nil {
		return []byte(""), fmt.Errorf("unable to decode returned data - %s", err)
	}

	return d, nil
}
