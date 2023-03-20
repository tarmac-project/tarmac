package logger

import (
	"testing"
)

type LoggerTestCase struct {
	name     string
	level    string
	message  string
	expected string
}

func TestLogger(t *testing.T) {
	// Define test cases
	tt := []LoggerTestCase{
		{
			name:     "Test Trace",
			level:    "trace",
			message:  "This is a trace log",
			expected: "This is a trace log",
		},
		{
			name:     "Test Debug",
			level:    "debug",
			message:  "This is a debug log",
			expected: "This is a debug log",
		},
		{
			name:     "Test Info",
			level:    "info",
			message:  "This is an info log",
			expected: "This is an info log",
		},
		{
			name:     "Test Error",
			level:    "error",
			message:  "This is an error log",
			expected: "This is an error log",
		},
	}

	// Run test cases
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize logger with mock hostCall
			logs := make(map[string]string)
			logger, err := New(Config{Namespace: "default", HostCall: func(namespace string, capability string, function string, input []byte) ([]byte, error) {
				if namespace != "default" || capability != "logger" || function != tc.level {
					t.Errorf("hostcall signature invalid %s, %s, %s", namespace, capability, function)
				}
				logs[function] = string(input)
				return []byte(""), nil
			}})
			if err != nil {
				t.Errorf("unexpected error initializing logger - %s", err)
			}

			// Call logger method
			switch tc.level {
			case "trace":
				logger.Trace(tc.message)
			case "debug":
				logger.Debug(tc.message)
			case "info":
				logger.Info(tc.message)
			case "error":
				logger.Error(tc.message)
			}

			// Check result
			v, ok := logs[tc.level]
			if !ok {
				t.Errorf("log message host call was not executed")
			}

			if v != tc.expected {
				t.Errorf("log was executed, but message is invalid - %s", v)
			}
		})
	}
}
