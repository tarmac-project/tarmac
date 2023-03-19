package sdk

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

// newSQL returns a new instance of SQL initalized with Config.
func newSQL(cfg Config) *SQL {
	return &SQL{namespace: cfg.Namespace, hostCall: cfg.hostCall}
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
		return []byte(""), fmt.Errorf("unexpected response from hostcall, did not contain data return")
	}

	// Decode SQL Response
	d, err := base64.StdEncoding.DecodeString(fastjson.GetString(rsp, "data"))
	if err != nil {
		return []byte(""), fmt.Errorf("unable to decode returned data - %s", err)
	}

	return d, nil
}
