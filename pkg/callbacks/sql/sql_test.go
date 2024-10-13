package sql

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	// Import MySQL Driver
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/tarmac-project/tarmac"

	proto "github.com/tarmac-project/protobuf-go/sdk/sql"
	pb "google.golang.org/protobuf/proto"
)

type TestCase struct {
	name string
	q    string
}

func TestInterface(t *testing.T) {
	_, err := New(Config{})
	if err == nil {
		t.Fatalf("New should error if DB is not provided")
	}
}

func TestSQLQuery(t *testing.T) {
	// Create a DB connection using in-memory SQLLite
	mockDB, err := sql.Open("mysql", "root:example@tcp(mysql:3306)/example")
	if err != nil {
		t.Fatalf("Unable to create sqllite DB for testing")
	}
	mockDB.SetMaxOpenConns(1)
	defer mockDB.Close()

	// Create new SQL Database instance
	db, err := New(Config{DB: mockDB})
	if err != nil {
		t.Fatalf("Unable to create new Database instance - %s", err)
	}

	// Setup Unhappy path tests
	var tc []TestCase
	tc = append(tc, TestCase{
		name: "Invalid Syntax",
		q:    `CREATE TALBE;`,
	})

	tc = append(tc, TestCase{
		name: "Empty Query",
		q:    "",
	})

	tc = append(tc, TestCase{
		name: "Table does not exist",
		q:    `SELECT * FROM nonexistanttablethatdoesnotexist;`,
	})

	tc = append(tc, TestCase{
		name: "Failing Query",
		q:    `1234e213423ewqw`,
	})

	t.Run("Protobuf Based Queries", func(t *testing.T) {

		t.Run("Unhappy Path", func(t *testing.T) {
			for _, c := range tc {
				t.Run(c.name, func(t *testing.T) {
					query := &proto.SQLQuery{Query: []byte(c.q)}
					qMsg, err := pb.Marshal(query)
					if err != nil {
						t.Fatalf("Unable to marshal query message")
					}

					r, err := db.Query(qMsg)
					if err == nil {
						t.Fatalf("Unexpected success with failure test case")
					}

					// Unmarshal the Tarmac Response
					var rsp proto.SQLQueryResponse
					err = pb.Unmarshal(r, &rsp)
					if err != nil {
						t.Fatalf("Error parsing returned query response")
					}

					// Check Status Codes
					if rsp.Status.Code == 200 {
						t.Fatalf("Unexpected Success with unhappy path test - %d", rsp.Status.Code)
					}
				})
			}
		})

		t.Run("Happy Path", func(t *testing.T) {
			t.Run("Create Table", func(t *testing.T) {
				query := &proto.SQLExec{Query: []byte(`CREATE TABLE IF NOT EXISTS testpkg ( id int NOT NULL AUTO_INCREMENT, name varchar(255), PRIMARY KEY (id) );`)}
				qMsg, err := pb.Marshal(query)
				if err != nil {
					t.Fatalf("Unable to marshal query message")
				}

				r, err := db.Query(qMsg)
				if err != nil {
					t.Fatalf("Unable to execute table creation query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp proto.SQLExecResponse
				err = pb.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}
			})

			t.Run("Insert Data", func(t *testing.T) {
				query := &proto.SQLExec{Query: []byte(`INSERT INTO testpkg (name)  VALUES ("John Smith");`)}
				qMsg, err := pb.Marshal(query)
				if err != nil {
					t.Fatalf("Unable to marshal query message")
				}

				r, err := db.Exec(qMsg)
				if err != nil {
					t.Errorf("Unable to execute INSERT query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp proto.SQLExecResponse
				err = pb.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}

				// Check Rows Affected
				if rsp.RowsAffected != 1 {
					t.Errorf("Unexpected rows affected - %d", rsp.RowsAffected)
				}

				// Check Last Insert ID
				if rsp.LastInsertId != 1 {
					t.Errorf("Unexpected last insert ID - %d", rsp.LastInsertId)
				}
			})

			t.Run("Select Data", func(t *testing.T) {
				query := &proto.SQLQuery{Query: []byte(`SELECT * from testpkg;`)}
				qMsg, err := pb.Marshal(query)
				if err != nil {
					t.Fatalf("Unable to marshal query message")
				}

				r, err := db.Query(qMsg)
				if err != nil {
					t.Errorf("Unable to execute SELECT query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp proto.SQLQueryResponse
				err = pb.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}

				// Verify Columns
				if len(rsp.Columns) != 2 {
					t.Fatalf("Unexpected number of columns returned - %d", len(rsp.Columns))
				}

				// Check Column Names
				if rsp.Columns[0] != "id" || rsp.Columns[1] != "name" {
					t.Fatalf("Unexpected column names returned - %v", rsp.Columns)
				}

				// Verify Data
				if len(rsp.Data) == 0 {
					t.Fatalf("No data returned from query")
				}

			})

			t.Run("Delete Table", func(t *testing.T) {
				query := &proto.SQLExec{Query: []byte(`DROP TABLE IF EXISTS testpkg;`)}
				qMsg, err := pb.Marshal(query)
				if err != nil {
					t.Fatalf("Unable to marshal query message")
				}

				r, err := db.Query(qMsg)
				if err != nil {
					t.Errorf("Unable to execute DELETE TABLE query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp proto.SQLExecResponse
				err = pb.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}
			})

		})
	})

	t.Run("Test SQLExec", func(t *testing.T) {
		t.Run("Happy Path", func(t *testing.T) {
			query := &proto.SQLExec{Query: []byte(`CREATE TABLE IF NOT EXISTS testpkg ( id int NOT NULL AUTO_INCREMENT, name varchar(255), PRIMARY KEY (id) );`)}
			qMsg, err := pb.Marshal(query)
			if err != nil {
				t.Fatalf("Unable to marshal query message")
			}

			r, err := db.Exec(qMsg)
			if err != nil {
				t.Fatalf("Unable to execute table creation query - %s", err)
			}

			// Unmarshal the Tarmac Response
			var rsp proto.SQLExecResponse
			err = pb.Unmarshal(r, &rsp)
			if err != nil {
				t.Fatalf("Error parsing returned query response")
			}

			// Check Status Codes
			if rsp.Status.Code != 200 {
				t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
			}
		})

		t.Run("Unhappy Path", func(t *testing.T) {
			query := &proto.SQLExec{Query: []byte(`CREATE TALBE;`)}
			qMsg, err := pb.Marshal(query)
			if err != nil {
				t.Fatalf("Unable to marshal query message")
			}

			r, err := db.Exec(qMsg)
			if err == nil {
				t.Fatalf("Unexpected success with failure test case")
			}

			// Unmarshal the Tarmac Response
			var rsp proto.SQLExecResponse
			err = pb.Unmarshal(r, &rsp)
			if err != nil {
				t.Fatalf("Error parsing returned query response")
			}

			// Check Status Codes
			if rsp.Status.Code == 200 {
				t.Fatalf("Unexpected Success with unhappy path test - %d", rsp.Status.Code)
			}
		})

		t.Run("Empty Exec", func(t *testing.T) {
			query := &proto.SQLExec{}
			qMsg, err := pb.Marshal(query)
			if err != nil {
				t.Fatalf("Unable to marshal query message")
			}

			r, err := db.Exec(qMsg)
			if err == nil {
				t.Fatalf("Unexpected success with failure test case")
			}

			// Unmarshal the Tarmac Response
			var rsp proto.SQLExecResponse
			err = pb.Unmarshal(r, &rsp)
			if err != nil {
				t.Fatalf("Error parsing returned query response")
			}

			// Check Status Codes
			if rsp.Status.Code == 200 {
				t.Fatalf("Unexpected Success with unhappy path test - %d", rsp.Status.Code)
			}
		})
	})

	// Test the JSON Interface for Backwards Compatibility
	t.Run("JSON Based Queries", func(t *testing.T) {

		t.Run("Unhappy Path", func(t *testing.T) {
			for _, c := range tc {
				t.Run(c.name, func(t *testing.T) {
					r, err := db.Query([]byte(fmt.Sprintf(`{"query":"%s"}`, base64.StdEncoding.EncodeToString([]byte(c.q)))))
					if err == nil {
						t.Fatalf("Unexpected success with failure test case")
					}

					// Unmarshal the Tarmac Response
					var rsp tarmac.SQLQueryResponse
					err = ffjson.Unmarshal(r, &rsp)
					if err != nil {
						t.Fatalf("Error parsing returned query response")
					}

					// Check Status Codes
					if rsp.Status.Code == 200 {
						t.Fatalf("Unexpected Success with unhappy path test - %v", rsp)
					}
				})
			}
		})

		t.Run("Bad JSON", func(t *testing.T) {
			_, err := db.Query([]byte(`{asdfas`))
			if err == nil {
				t.Fatalf("Unexpected success with bad input")
			}
		})

		t.Run("Bad Base64", func(t *testing.T) {
			_, err := db.Query([]byte(`{"query": "my bologna has a first name it's this is not base64...."}`))
			if err == nil {
				t.Fatalf("Unexpected success with bad input")
			}
		})

		t.Run("Happy Path", func(t *testing.T) {
			t.Run("Create Table", func(t *testing.T) {
				q := base64.StdEncoding.EncodeToString([]byte(`CREATE TABLE IF NOT EXISTS testpkg ( id int NOT NULL, name varchar(255), PRIMARY KEY (id) );`))
				j := fmt.Sprintf(`{"query":"%s"}`, q)
				r, err := db.Query([]byte(j))
				if err != nil {
					t.Fatalf("Unable to execute table creation query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp tarmac.SQLQueryResponse
				err = ffjson.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}
			})

			t.Run("Insert Data", func(t *testing.T) {
				q := base64.StdEncoding.EncodeToString([]byte(`INSERT INTO testpkg (id, name)  VALUES (1, "John Smith");`))
				j := fmt.Sprintf(`{"query":"%s"}`, q)
				r, err := db.Query([]byte(j))
				if err != nil {
					t.Errorf("Unable to execute table creation query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp tarmac.SQLQueryResponse
				err = ffjson.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}
			})

			t.Run("Select Data", func(t *testing.T) {
				q := base64.StdEncoding.EncodeToString([]byte(`SELECT * from testpkg;`))
				j := fmt.Sprintf(`{"query":"%s"}`, q)
				r, err := db.Query([]byte(j))
				if err != nil {
					t.Errorf("Unable to execute table creation query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp tarmac.SQLQueryResponse
				err = ffjson.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}

				// Verify query response
				data, err := base64.StdEncoding.DecodeString(rsp.Data)
				if err != nil {
					t.Fatalf("Callback returned undecodable response - %s", err)
				}

				// Parse returned SQL Data
				type rowData struct {
					ID   []byte `json:"id"`
					Name []byte `json:"name"`
				}

				var records []rowData
				err = ffjson.Unmarshal(data, &records)
				if err != nil {
					t.Fatalf("Unable to unmarshal SQL response - %s", err)
				}

				id, err := strconv.Atoi(string(records[0].ID))
				if err != nil {
					t.Fatalf("Unable to convert ID to integer - %s", err)
				}

				if id != 1 {
					t.Fatalf("Unexpected value from Database got %d", id)
				}

			})

			t.Run("Delete Table", func(t *testing.T) {
				q := base64.StdEncoding.EncodeToString([]byte(`DROP TABLE IF EXISTS testpkg;`))
				j := fmt.Sprintf(`{"query":"%s"}`, q)
				r, err := db.Query([]byte(j))
				if err != nil {
					t.Errorf("Unable to execute table creation query - %s", err)
				}

				// Unmarshal the Tarmac Response
				var rsp tarmac.SQLQueryResponse
				err = ffjson.Unmarshal(r, &rsp)
				if err != nil {
					t.Fatalf("Error parsing returned query response")
				}

				// Check Status Codes
				if rsp.Status.Code != 200 {
					t.Fatalf("Callback execution did not work, returned %d - %s", rsp.Status.Code, rsp.Status.Status)
				}
			})
		})
	})
}
