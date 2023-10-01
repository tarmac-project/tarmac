package sql

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	// Import MySQL Driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/tarmac-project/tarmac"
	"github.com/pquerna/ffjson/ffjson"
	"testing"
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

	t.Run("Unhappy Path", func(t *testing.T) {
		for _, c := range tc {
			t.Run(c.name, func(t *testing.T) {
				r, err := db.Query([]byte(fmt.Sprintf(`{"query":"%s"}`, base64.StdEncoding.EncodeToString([]byte(c.q)))))
				if err == nil {
					t.Fatalf("Unexepected success with failure test case")
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
				t.Logf("Returned Status %d - %s", rsp.Status.Code, rsp.Status.Status)
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
}
