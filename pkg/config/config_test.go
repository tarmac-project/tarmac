package config

import (
	"os"
	"testing"
)

type TestCase struct {
	name  string
	data  []byte
	valid bool
}

// TestParserFile is a unit test for the Parse function
func TestParserFile(t *testing.T) {
	// Define the JSON data that will be used to create the temporary file
	data := []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"http","path":"/example","methods":["GET"],"function":"example"}]}}}`)

	// Create a temporary file in the /tmp directory
	fh, err := os.CreateTemp("/tmp", "")
	if err != nil {
		// If there is an error creating the temporary file, fail the test with the error message
		t.Fatalf("unable to create temp file: %s", err)
	}

	// Ensure the file handle is closed and the temporary file is removed at the end of the function
	defer fh.Close()
	defer os.Remove(fh.Name())

	// Get the filename of the temporary file
	fn := fh.Name()

	// Write the JSON data to the temporary file
	_, err = fh.Write(data)
	if err != nil {
		// If there is an error writing to the temporary file, fail the test with the error message
		t.Fatalf("could not write temp file: %s", err)
	}

	// Close the temporary file
	err = fh.Close()
	if err != nil {
		// If there is an error closing the temporary file, fail the test with the error message
		t.Fatalf("could not close temp file: %s", err)
	}

	// Parse the temporary file
	_, err = Parse(fn)
	if err != nil {
		// If there is an error parsing the file, fail the test with the error message
		t.Errorf("could not parse file: %s", err)
	}
}

func TestMissingFile(t *testing.T) {
	_, err := Parse("/tmp/doesnotexist.nope")
	if err == nil {
		t.Errorf("expected file not found, did not get error")
	}
}

// This is a unit test function for the Parse function
func TestConfigParser(t *testing.T) {
	// Define an array of test cases
	tt := []TestCase{
		{
			name:  "Empty JSON",
			data:  []byte(``),
			valid: false,
		},
		{
			name:  "Invalid JSON Format",
			data:  []byte(`{invalid_json`),
			valid: false,
		},
		{
			name:  "Valid JSON",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"http","path":"/example","methods":["GET"],"function":"example"}]}}}`),
			valid: true,
		},
	}

	// Loop through the test cases
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file in the /tmp directory
			fh, err := os.CreateTemp("/tmp", "")
			if err != nil {
				// If there is an error creating the temporary file, fail the test with the error message
				t.Fatalf("unable to create temp file: %s", err)
			}

			// Ensure the file handle is closed and the temporary file is removed at the end of the function
			defer fh.Close()
			defer os.Remove(fh.Name())

			// Get the filename of the temporary file
			fn := fh.Name()

			// Write the JSON data to the temporary file
			_, err = fh.Write(tc.data)
			if err != nil {
				// If there is an error writing to the temporary file, fail the test with the error message
				t.Fatalf("could not write temp file: %s", err)
			}

			// Close the temporary file
			err = fh.Close()
			if err != nil {
				// If there is an error closing the temporary file, fail the test with the error message
				t.Fatalf("could not close temp file: %s", err)
			}

			// Parse the temporary file
			_, err = Parse(fn)

			// Check the error returned by Parse against the expected value for the test case
			if err != nil && tc.valid {
				t.Errorf("could not parse file: %s", err)
			}
			if err == nil && !tc.valid {
				t.Errorf("did not get expected error parsing file")
			}
		})
	}
}
