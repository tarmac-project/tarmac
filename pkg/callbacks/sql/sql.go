/*
Package database is part of the Tarmac suite of Host Callback packages. This package provides users with the ability to
provide WASM functions with a host callback interface that provides SQL database capabilities.

	import (
		"github.com/tarmac-project/tarmac/pkg/callbacks/sql"
	)

	func main() {
		// Create instance of database to register for callback execution
		dBase := database.New(sql.Config{})

		// Create Callback router and register
		router := callbacks.New()
		router.RegisterCallback("sql", "query", dBase.Query)
	}
*/
package sql

import (
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/tarmac-project/tarmac"

	"github.com/tarmac-project/protobuf-go/sdk"
	proto "github.com/tarmac-project/protobuf-go/sdk/sql"
	pb "google.golang.org/protobuf/proto"
)

// Database provides access to Host Callbacks that interface with a SQL database within Tarmac. The callbacks
// within Database provide all the logic and error hangling of accessing and interacting with a SQL database.
// Users will send the specified JSON request to execute a query and receive an appropriate JSON response.
type Database struct {
	db *sql.DB
}

// Config is provided to users to configure the Host Callback. All Tarmac Callbacks follow the same configuration
// format; each Config struct gives the specific Host Callback unique functionality.
type Config struct {
	// DB is the user-provided SQL database instance using the standard "database/sql" interface. This package by
	// itself does not manage database connections but rather relies on the sql.DB interface. Users must
	// supply an initiated sql.DB to work with.
	DB *sql.DB
}

// New will create and return a new Database instance that users can register as a Tarmac Host Callback function. Users
// can provide any custom Database configurations using the configuration options supplied.
func New(cfg Config) (*Database, error) {
	db := &Database{}
	if cfg.DB == nil {
		return db, fmt.Errorf("DB cannot be nil")
	}
	db.db = cfg.DB
	return db, nil
}

// Exec will execute the supplied query against the supplied database. Error handling, processing results, and base64 encoding
// of data are all handled via this function. Note, this function expects the SQLExec type as input
// and will return a SQLExecResponse JSON.
func (db *Database) Exec(b []byte) ([]byte, error) {
	msg := &proto.SQLExec{}
	err := pb.Unmarshal(b, msg)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to unmarshal database:exec request")
	}

	r := &proto.SQLExecResponse{}
	r.Status = &sdk.Status{Code: 200, Status: "OK"}

	if len(msg.Query) < 1 {
		r.Status.Code = 400
		r.Status.Status = "SQL Query must be defined"
	}

	var results sql.Result
	if r.Status.Code == 200 {
		results, err = db.exec(msg.Query)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to execute query - %s", err)
		}
	}

	if r.Status.Code == 200 {
		// Set Row Count
		ra, err := results.RowsAffected()
		if err != nil {
			r.Status.Code = 206
			r.Status.Status = fmt.Sprintf("Unable to get rows affected - %s", err)
			r.RowsAffected = 0
		}
		r.RowsAffected = ra

		// Set Last Insert ID
		id, err := results.LastInsertId()
		if err != nil {
			r.Status.Code = 206
			r.Status.Status = fmt.Sprintf("Unable to get last insert ID - %s", err)
			r.LastInsertId = 0
		}
		r.LastInsertId = id
	}

	rsp, err := pb.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal database:exec response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}

	return rsp, fmt.Errorf("%s", r.Status.Status)
}

// Query will execute the supplied query against the supplied database. Error handling, processing results, and base64 encoding
// of data are all handled via this function. Note, this function expects the SQLQueryJSON type as input
// and will return a SQLQueryResponse JSON.
func (db *Database) Query(b []byte) ([]byte, error) {
	msg := &proto.SQLQuery{}
	err := pb.Unmarshal(b, msg)
	if err != nil {
		// Fallback to JSON if proto fails
		return db.queryJSON(b)
	}

	// Create a new SQLQueryResponse
	r := &proto.SQLQueryResponse{}
	r.Status = &sdk.Status{Code: 200, Status: "OK"}

	if len(msg.Query) < 1 {
		r.Status.Code = 400
		r.Status.Status = "SQL Query must be defined"
	}

	if r.Status.Code == 200 {
		columns, results, err := db.query(msg.Query)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to execute query - %s", err)
		}

		// Marshal results into JSON bytes
		j, err := ffjson.Marshal(results)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to convert query results to JSON - %s", err)
		}

		// Set the response data
		r.Data = j
		r.Columns = columns
	}

	// Marshal a response Proto to return to caller
	rsp, err := pb.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal database:query response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}

	return rsp, fmt.Errorf("%s", r.Status.Status)
}

// queryJSON retains the JSON interface for backwards compatibility with the Tarmac Host Callback interface.
func (db *Database) queryJSON(b []byte) ([]byte, error) {
	// Start Response Message assuming everything is good
	r := tarmac.SQLQueryResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	// Parse incoming Request
	var rq tarmac.SQLQuery
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = "Error Parsing Input"
	}

	// Decode Query to execute
	q, err := base64.StdEncoding.DecodeString(rq.Query)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = fmt.Sprintf("Unable to decode query - %s", err)
	}

	if len(q) < 1 {
		r.Status.Code = 400
		r.Status.Status = "SQL Query must be defined"
	}

	if r.Status.Code == 200 {
		_, results, err := db.query(q)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to execute query - %s", err)
		}

		// Convert results into JSON
		j, err := ffjson.Marshal(results)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to convert query results to JSON - %s", err)
		}

		// Base64 encode results to avoid JSON format conflicts
		r.Data = base64.StdEncoding.EncodeToString(j)
	}

	// Marshal a resposne JSON to return to caller
	rsp, err := ffjson.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal database:query response")
	}

	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}

// query will execute the supplied query against the database and return
// the rows as a list of maps. The keys in the map are the column names
// and the values are the column values.
func (db *Database) query(qry []byte) ([]string, []map[string]any, error) {

	rows, err := db.db.Query(string(qry))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to execute query - %s", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to process query results - %s", err)
	}

	var results []map[string]any

	for rows.Next() {
		colNames := make([]interface{}, len(columns))
		data := make([]interface{}, len(columns))
		for i := range colNames {
			data[i] = &colNames[i]
		}

		err := rows.Scan(data...)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to process query results - %s", err)
		}

		m := make(map[string]any)
		for i, c := range columns {
			val := *data[i].(*interface{})
			m[c] = val
		}

		results = append(results, m)
	}

	if rows.Err() != nil {
		return nil, nil, fmt.Errorf("error while processing query results - %s", rows.Err())
	}

	return columns, results, nil
}

// exec will execute the supplied query against the database and return the result.
func (db *Database) exec(qry []byte) (sql.Result, error) {
	return db.db.Exec(string(qry))
}
