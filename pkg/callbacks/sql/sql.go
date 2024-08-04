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

	"github.com/tarmac-project/tarmac/proto"
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

	r := proto.SQLQueryResponse{}
	r.Status = &proto.Status{Code: 200, Status: "OK"}

	if len(msg.Query) < 1 {
		r.Status.Code = 400
		r.Status.Status = "SQL Query must be defined"
	}

	if r.Status.Code == 200 {
		rows, err := db.db.Query(msg.Query)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to execute query - %s", err)
		}
		defer rows.Close()

		if r.Status.Code == 200 {

			// Set last insert ID
			lastID, err := rows.LastInsertId()
			if err != nil {
				r.Status.Code = 500
				r.Status.Status = fmt.Sprintf("Unable to get last insert ID - %s", err)
			}

			if r.Status.Code == 200 {
				r.LastInsertID = lastID
			}

			// Set number of rows affected
			rowsAffected, err := rows.RowsAffected()
			if err != nil {
				r.Status.Code = 500
				r.Status.Status = fmt.Sprintf("Unable to get rows affected - %s", err)
			}

			if r.Status.Code == 200 {
				r.RowsAffected = rowsAffected
			}

			// Grab Colummns
			columns, err := rows.Columns()
			if err != nil {
				r.Status.Code = 500
				r.Status.Status = fmt.Sprintf("Unable to process query results - %s", err)
			}
			r.Columns = columns

			// Loop through results
			if len(columns) > 0 {
				pbRows := []*proto.Row{}
				for rows.Next() {
					pbRow := proto.Row{}
					data := make(map[string][]byte)
					rawdata := make([]*sql.RawBytes, len(columns))
					for i := range columns {
						rawdata[i] = new(sql.RawBytes)
					}

					err := rows.Scan(rawdata...)
					if err != nil {
						r.Status.Code = 500
						r.Status.Status = fmt.Sprintf("Unable to process query results - %s", err)
					}

					for i, raw := range rawdata {
						if raw != nil {
							data[columns[i]] = *raw
						}
					}

					if r.Status.Code == 200 {
						pbRow.Data = data
						pbRows = append(pbRows, &pbRow)
					}
				}
				r.Rows = pbRows
			}
		}

		// Marshal a resposne
		rsp, err := pb.Marshal(&r)
		if err != nil {
			return []byte(""), fmt.Errorf("unable to marshal database:query response")
		}

		// Return response to caller
		if r.Status.Code == 200 {
			return rsp, nil
		}
		return rsp, fmt.Errorf("%s", r.Status.Status)
	}
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
		var results []map[string]interface{}

		// Query database
		rows, err := db.db.Query(string(q))
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to execute query - %s", err)
		}

		if r.Status.Code == 200 {
			defer rows.Close()

			// Grab column details for result processing
			columns, err := rows.ColumnTypes()
			if err != nil {
				r.Status.Code = 500
				r.Status.Status = fmt.Sprintf("Unable to process query results - %s", err)
			}

			if len(columns) > 0 {
				// Loop through results creating a list of maps
				for rows.Next() {
					colNames := make([]interface{}, len(columns))
					data := make([]interface{}, len(columns))
					for i := range colNames {
						data[i] = &colNames[i]
					}

					// Extract data from results
					err := rows.Scan(data...)
					if err != nil {
						r.Status.Code = 500
						r.Status.Status = fmt.Sprintf("Unable to process query results - %s", err)
					}

					// Create a map for simple access to data
					m := make(map[string]interface{})
					for i, c := range columns {
						m[c.Name()] = data[i]
					}

					// Append to final results
					results = append(results, m)
				}
				if rows.Err() != nil {
					r.Status.Code = 500
					r.Status.Status = fmt.Sprintf("Error while processing query results - %s", rows.Err())
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
		}
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
