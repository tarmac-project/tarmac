package sql

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSQL_Query(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		query     string
		hostCall  func(string, string, string, []byte) ([]byte, error)
		expected  []byte
		err       bool
	}{
		{
			name:      "success",
			namespace: "test-namespace",
			query:     "SELECT * FROM users;",
			hostCall: func(namespace, service, endpoint string, payload []byte) ([]byte, error) {
				// Validate hostcall parameters are as expected
				if namespace != "test-namespace" || service != "sql" || endpoint != "query" {
					return nil, fmt.Errorf("unexpected arguments - namespace: %s, service: %s, endpoint: %s", namespace, service, endpoint)
				}

				// Validate Payload
				expectedPayload := []byte(`{"query":"U0VMRUNUICogRlJPTSB1c2Vyczs="}`) // "SELECT * FROM users;" base64 encoded
				if !bytes.Equal(payload, expectedPayload) {
					return nil, fmt.Errorf("unexpected payload - got: %s, expected: %s", payload, expectedPayload)
				}

				// Return valid JSON base64 encoded
				return []byte(`{"data":"eyJ1c2VycyI6W3sidXNlciI6IjEiLCJuYW1lIjoiSm9obiBEb2UiLCJhZGRyZXNzIjoiMTIzIFN0cmVldGNvcm5lciJ9XX0="}`), nil
			},
			expected: []byte(`{"users":[{"user":"1","name":"John Doe","address":"123 Streetcorner"}]}`),
			err:      false,
		},
		{
			name:      "hostcall error",
			namespace: "test-namespace",
			query:     "SELECT * FROM users;",
			hostCall: func(namespace, service, endpoint string, payload []byte) ([]byte, error) {
				return []byte(""), fmt.Errorf("an error")
			},
			expected: nil,
			err:      true,
		},
		{
			name:      "decode error",
			namespace: "test-namespace",
			query:     "SELECT * FROM users;",
			hostCall: func(namespace, service, endpoint string, payload []byte) ([]byte, error) {
				return []byte(`{"data":"!@#$%^&*"}`), nil
			},
			expected: nil,
			err:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize SQL instance
			sql, err := New(Config{Namespace: tc.namespace, HostCall: tc.hostCall})
			if err != nil {
				t.Errorf("Unexpected error while creating SQL instance - %s", err)
			}

			// Call Query
			rsp, err := sql.Query(tc.query)
			if err != nil && tc.err {
				return
			}
			if err != nil && !tc.err {
				t.Errorf("Unexpected error when calling SQL Query - %s", err)
			}
			if err == nil && tc.err {
				t.Errorf("Expected error when calling SQL Query got nil")
			}

			// Validate response is as expected
			if !bytes.Equal(rsp, tc.expected) {
				t.Errorf("Did not get expected response from function - %s", err)
			}
		})
	}
}
