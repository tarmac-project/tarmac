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

func TestParserFile(t *testing.T) {
	// Define the JSON data that will be used to create the temporary file
	data := []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm","pool_size":10},"example-defaults":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"http","path":"/example","methods":["GET"],"function":"example"},{"type":"scheduled_task","function":"example","frequency":15},{"type":"init","function":"example","retries":15,"frequency":50},{"type":"init","function":"example-defaults"},{"type":"function","function":"function1"}]}}}`)

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
	cfg, err := Parse(fn)
	if err != nil {
		// If there is an error parsing the file, fail the test with the error message
		t.Errorf("could not parse file: %s", err)
	}

	// Validate Correct Values
	t.Run("Validate Correct Values", func(t *testing.T) {
		t.Run("Validate Service Name", func(t *testing.T) {
			if cfg.Services["example"].Name != "example" {
				t.Errorf("Unexpected Service Name - %s", cfg.Services["example"].Name)
			}
		})

		t.Run("Validate Function Filepath", func(t *testing.T) {
			if cfg.Services["example"].Functions["example"].Filepath != "./functions/example.wasm" {
				t.Errorf("Unexpected Function Filepath - %s", cfg.Services["example"].Functions["example"].Filepath)
			}
		})

		t.Run("Validate Function Pool Size", func(t *testing.T) {
			if cfg.Services["example"].Functions["example"].PoolSize != 10 {
				t.Errorf("Unexpected Function Pool Size - %d", cfg.Services["example"].Functions["example"].PoolSize)
			}
		})

		t.Run("Validate Default Pool Size", func(t *testing.T) {
			if cfg.Services["example"].Functions["example-defaults"].PoolSize != 100 {
				t.Errorf("Unexpected Default Pool Size - %d", cfg.Services["example"].Functions["example-defaults"].PoolSize)
			}
		})

		t.Run("Validate Route Type", func(t *testing.T) {
			if cfg.Services["example"].Routes[0].Type != "http" {
				t.Errorf("Unexpected Route Type - %s", cfg.Services["example"].Routes[0].Type)
			}

			if cfg.Services["example"].Routes[1].Type != "scheduled_task" {
				t.Errorf("Unexpected Route Type - %s", cfg.Services["example"].Routes[1].Type)
			}

			if cfg.Services["example"].Routes[2].Type != "init" {
				t.Errorf("Unexpected Route Type - %s", cfg.Services["example"].Routes[2].Type)
			}

			if cfg.Services["example"].Routes[3].Type != "init" {
				t.Errorf("Unexpected Route Type - %s", cfg.Services["example"].Routes[3].Type)
			}

			if cfg.Services["example"].Routes[4].Type != "function" {
				t.Errorf("Unexpected Route Type - %s", cfg.Services["example"].Routes[4].Type)
			}
		})

		t.Run("Validate Route Path", func(t *testing.T) {
			if cfg.Services["example"].Routes[0].Path != "/example" {
				t.Errorf("Unexpected Route Path - %s", cfg.Services["example"].Routes[0].Path)
			}
		})

		t.Run("Validate Route Methods", func(t *testing.T) {
			if cfg.Services["example"].Routes[0].Methods[0] != "GET" {
				t.Errorf("Unexpected Route Method - %s", cfg.Services["example"].Routes[0].Methods[0])
			}
		})

		t.Run("Validate Route Function", func(t *testing.T) {
			if cfg.Services["example"].Routes[0].Function != "example" {
				t.Errorf("Unexpected Route Function - %s", cfg.Services["example"].Routes[0].Function)
			}
		})

		t.Run("Validate Scheduled Task Frequency", func(t *testing.T) {
			if cfg.Services["example"].Routes[1].Frequency != 15 {
				t.Errorf("Unexpected Scheduled Task Frequency - %d", cfg.Services["example"].Routes[1].Frequency)
			}
		})

		t.Run("Validate Init Retries", func(t *testing.T) {
			if cfg.Services["example"].Routes[2].Retries != 15 {
				t.Errorf("Unexpected Init Retries - %d", cfg.Services["example"].Routes[2].Retries)
			}
		})

		t.Run("Validate Init Frequency", func(t *testing.T) {
			if cfg.Services["example"].Routes[2].Frequency != 50 {
				t.Errorf("Unexpected Init Frequency - %d", cfg.Services["example"].Routes[2].Frequency)
			}
		})

		t.Run("Validate Default Init Frequency", func(t *testing.T) {
			if cfg.Services["example"].Routes[3].Frequency != 1 {
				t.Errorf("Unexpected Default Init Frequency - %d", cfg.Services["example"].Routes[3].Frequency)
			}
		})

		t.Run("Validate Function Name", func(t *testing.T) {
			if cfg.Services["example"].Routes[4].Function != "function1" {
				t.Errorf("Unexpected Function Name - %s", cfg.Services["example"].Routes[4].Function)
			}
		})

	})

	// Validate RouteLookup
	t.Run("Validate RouteLookup with Valid Route", func(t *testing.T) {
		_, err := cfg.RouteLookup("http:GET:/example")
		if err != nil {
			t.Errorf("Unexpected failure looking up route")
		}
	})

	t.Run("Validate RouteLookup with Invalid Route", func(t *testing.T) {
		_, err := cfg.RouteLookup("http:DELETE:/example")
		if err == nil {
			t.Errorf("Unexpected success looking up routei - %s", err)
		}
	})
}

func TestMissingFile(t *testing.T) {
	_, err := Parse("/tmp/doesnotexist.nope")
	if err == nil {
		t.Errorf("expected file not found, did not get error")
	}
}

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
		{
			name:  "Missing Frequency for scheduled_task",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"scheduled_task","function":"default"},{"type":"init","function":"default"},{"type":"function","function":"function1"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Function for scheduled_task",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"scheduled_task","frequency":15},{"type":"init","function":"default"},{"type":"function","function":"function1"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Function for init",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"scheduled_task","function":"default","frequency":15},{"type":"init"},{"type":"function","function":"function1"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Function for function",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"scheduled_task","function":"default","frequency":15},{"type":"init","function":"default"},{"type":"function"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Service Name",
			data:  []byte(`{"services":{"example":{"functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"scheduled_task","function":"default","frequency":15},{"type":"init","function":"default"},{"type":"function","function":"function1"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Function Name",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"scheduled_task","function":"default","frequency":15},{"type":"init","function":"default"},{"type":"function"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Function Filepath",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{},"routes":[{"type":"scheduled_task","function":"default","frequency":15},{"type":"init","function":"default"},{"type":"function","function":"function1"}]}}}}`),
			valid: false,
		},
		{
			name:  "Missing Route Type",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"path":"/example","methods":["GET"],"function":"example"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Route Path",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"http","methods":["GET"],"function":"example"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Route Methods",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"http","path":"/example","function":"example"}]}}}`),
			valid: false,
		},
		{
			name:  "Missing Route Function",
			data:  []byte(`{"services":{"example":{"name":"example","functions":{"example":{"filepath":"./functions/example.wasm"}},"routes":[{"type":"http","path":"/example","methods":["GET"]}]}}}`),
			valid: false,
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
