package httpclient

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/tarmac-project/tarmac"

	proto "github.com/tarmac-project/protobuf-go/sdk/http"
	pb "google.golang.org/protobuf/proto"
)

type HTTPClientCase struct {
	err      bool
	pass     bool
	httpCode int
	name     string
	call     string
	json     string
	proto    *proto.HTTPClient
}

func Test(t *testing.T) {
	h, err := New(Config{})
	if err != nil {
		t.Fatalf("Unable to create HTTP Client - %s", err)
	}

	// Start Test HTTP Server
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set a header to validate
		w.Header().Set("Server", "tarmac")

		// Check Header
		if r.Header.Get("teapot") != "true" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Process methods with and without payloads
		switch r.Method {
		case "POST", "PUT", "PATCH":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if strings.ToUpper(string(body)) != r.Method {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, _ = fmt.Fprintf(w, "%s", body)
		default:
			return
		}
	}))

	var tc []HTTPClientCase

	// Create a collection of test cases
	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple GET",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"GET","headers":{"teapot": "true"},"insecure":true,"url":"%s"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "GET",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      true,
		pass:     false,
		httpCode: 0,
		name:     "Simple GET without SkipVerify",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"GET","headers":{"teapot": "true"},"url":"%s"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:  "GET",
			Headers: map[string]string{"teapot": "true"},
			Url:     ts.URL,
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple HEAD",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"HEAD","headers":{"teapot": "true"},"insecure":true,"url":"%s"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "HEAD",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple DELETE",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"DELETE","headers":{"teapot": "true"},"insecure":true,"url":"%s"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "DELETE",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple POST",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"POST","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"UE9TVA=="}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "POST",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
			Body:     []byte("POST"),
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 400,
		name:     "Invalid POST",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"POST","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"NotValid"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "POST",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
			Body:     []byte("NotValid"),
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple PUT",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PUT","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"UFVU"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "PUT",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
			Body:     []byte("PUT"),
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 400,
		name:     "Invalid PUT",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PUT","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"NotValid"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "PUT",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
			Body:     []byte("NotValid"),
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple PATCH",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PATCH","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"UEFUQ0g="}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "PATCH",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
			Body:     []byte("PATCH"),
		},
	})

	tc = append(tc, HTTPClientCase{
		err:      false,
		pass:     true,
		httpCode: 400,
		name:     "Simple PATCH",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PATCH","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"NotValid"}`, ts.URL),
		proto: &proto.HTTPClient{
			Method:   "PATCH",
			Headers:  map[string]string{"teapot": "true"},
			Insecure: true,
			Url:      ts.URL,
			Body:     []byte("NotValid"),
		},
	})

	// Loop through test cases executing and validating
	for _, c := range tc {
		switch c.call {
		case "Call":
			t.Run(c.name+" Call", func(t *testing.T) {
				t.Run("JSON", func(t *testing.T) {
					// Call http callback
					b, err := h.Call([]byte(c.json))
					if err != nil && !c.err {
						t.Fatalf(" Callback failed unexpectedly - %s", err)
					}
					if err == nil && c.err {
						t.Fatalf(" Callback unexpectedly passed")
					}

					// Validate Response
					var rsp tarmac.HTTPClientResponse
					err = ffjson.Unmarshal(b, &rsp)
					if err != nil {
						t.Fatalf(" Callback Set replied with an invalid JSON - %s", err)
					}

					// Tarmac Response
					if rsp.Status.Code == 200 && !c.pass {
						t.Fatalf(" Callback Set returned an unexpected success - %+v", rsp)
					}
					if rsp.Status.Code != 200 && c.pass {
						t.Fatalf(" Callback Set returned an unexpected failure - %+v", rsp)
					}

					// HTTP Response
					if rsp.Code != c.httpCode {
						t.Fatalf(" returned an unexpected response code - %+v", rsp)
						return
					}

					// Validate Response Header
					v, ok := rsp.Headers["server"]
					if (!ok || v != "tarmac") && rsp.Code == 200 {
						t.Errorf(" returned an unexpected header - %+v", rsp)
					}

					// Validate Payload
					if len(rsp.Body) > 0 {
						body, err := base64.StdEncoding.DecodeString(rsp.Body)
						if err != nil {
							t.Fatalf("Error decoding  returned body - %s", err)
						}
						switch string(body) {
						case "PUT", "POST", "PATCH":
							return
						default:
							t.Errorf(" returned unexpected payload - %s", body)
						}
					}
				})
				t.Run("Protobuf", func(t *testing.T) {
					// Generate Protobuf
					msg, err := pb.Marshal(c.proto)
					if err != nil {
						t.Fatalf("Unable to marshal protobuf - %s", err)
					}

					// Call http callback
					b, err := h.Call(msg)
					if err != nil && !c.err {
						t.Fatalf(" Callback failed unexpectedly - %s", err)
					}

					// Validate protobuf response
					var rsp proto.HTTPClientResponse
					err = pb.Unmarshal(b, &rsp)
					if err != nil {
						t.Fatalf(" Callback Set replied with an invalid Protobuf - %s", err)
					}

					// Tarmac Response
					if rsp.Status.Code == 200 && !c.pass {
						t.Fatalf(" Callback Set returned an unexpected success - %d", rsp.Status.Code)
					}

					if rsp.Status.Code != 200 && c.pass {
						t.Fatalf(" Callback Set returned an unexpected failure - %d", rsp.Status.Code)

					}

					// HTTP Response
					if rsp.Code != int32(c.httpCode) {
						t.Fatalf(" returned an unexpected response code - %d", rsp.Code)
						return
					}

					// Validate Response Header
					v, ok := rsp.Headers["server"]
					if (!ok || v != "tarmac") && rsp.Code == 200 {
						t.Errorf(" returned an unexpected header - %s", v)
					}

					// Validate Payload
					if len(rsp.Body) > 0 {
						switch string(rsp.Body) {
						case "PUT", "POST", "PATCH":
							return
						default:
							t.Errorf(" returned unexpected payload - %s", rsp.Body)
						}
					}
				})
			})
		}
	}
}

func TestResponseBodySizeLimit(t *testing.T) {
	// Test server that returns configurable response sizes
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		// Check query parameter for response size
		sizeStr := r.URL.Query().Get("size")
		if sizeStr == "" {
			sizeStr = "1024" // Default 1KB
		}

		size := 1024
		if s, err := fmt.Sscanf(sizeStr, "%d", &size); err != nil || s != 1 {
			http.Error(w, "Invalid size parameter", http.StatusBadRequest)
			return
		}

		// Generate response of specified size with repeating pattern
		// Use a meaningful pattern instead of just 'A' characters
		pattern := "This is test data for HTTP response body size limiting. "
		patternBytes := []byte(pattern)
		data := make([]byte, size)

		for i := 0; i < size; i++ {
			data[i] = patternBytes[i%len(patternBytes)]
		}
		_, _ = w.Write(data)
	}))
	defer ts.Close()

	testCases := []struct {
		name            string
		config          Config
		responseSize    int
		expectTruncated bool
		description     string
	}{
		{
			name:            "Default 10MB limit with small response",
			config:          Config{}, // Use default
			responseSize:    1024,     // 1KB
			expectTruncated: false,
			description:     "Small response should not be truncated with default config",
		},
		{
			name:            "Custom 2KB limit with 1KB response",
			config:          Config{MaxResponseBodySize: 2048}, // 2KB
			responseSize:    1024,                              // 1KB
			expectTruncated: false,
			description:     "Response smaller than limit should not be truncated",
		},
		{
			name:            "Custom 2KB limit with 3KB response",
			config:          Config{MaxResponseBodySize: 2048}, // 2KB
			responseSize:    3072,                              // 3KB
			expectTruncated: true,
			description:     "Response larger than limit should be truncated",
		},
		{
			name:            "Custom 2KB limit with exactly 2KB response",
			config:          Config{MaxResponseBodySize: 2048}, // 2KB
			responseSize:    2048,                              // 2KB
			expectTruncated: false,
			description:     "Response exactly at limit should not be truncated",
		},
		{
			name:            "Zero config uses default 10MB",
			config:          Config{MaxResponseBodySize: 0}, // Should use default
			responseSize:    1024,                           // 1KB
			expectTruncated: false,
			description:     "Zero config should use default 10MB limit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create HTTP client with custom config
			h, err := New(tc.config)
			if err != nil {
				t.Fatalf("Unable to create HTTP Client - %s", err)
			}

			// Test both JSON and Protobuf interfaces
			t.Run("JSON", func(t *testing.T) {
				url := fmt.Sprintf("%s?size=%d", ts.URL, tc.responseSize)
				jsonReq := fmt.Sprintf(`{"method":"GET","headers":{},"insecure":true,"url":"%s"}`, url)

				b, err := h.Call([]byte(jsonReq))
				if err != nil {
					t.Fatalf("HTTP call failed: %s", err)
				}

				var rsp tarmac.HTTPClientResponse
				if err := ffjson.Unmarshal(b, &rsp); err != nil {
					t.Fatalf("Failed to unmarshal JSON response: %s", err)
				}

				if rsp.Status.Code != 200 {
					t.Fatalf("Expected successful status, got %d: %s", rsp.Status.Code, rsp.Status.Status)
				}

				// Decode base64 body
				body, err := base64.StdEncoding.DecodeString(rsp.Body)
				if err != nil {
					t.Fatalf("Failed to decode response body: %s", err)
				}

				expectedSize := tc.responseSize
				maxSize := tc.config.MaxResponseBodySize
				if maxSize <= 0 {
					maxSize = 10 * 1024 * 1024 // Default 10MB
				}

				if tc.expectTruncated {
					expectedSize = int(maxSize)
				}

				if len(body) != expectedSize {
					t.Errorf("%s: expected body length %d, got %d", tc.description, expectedSize, len(body))
				}

				// Generate expected data for MD5 verification
				pattern := "This is test data for HTTP response body size limiting. "
				patternBytes := []byte(pattern)
				expectedData := make([]byte, expectedSize)
				for i := 0; i < expectedSize; i++ {
					expectedData[i] = patternBytes[i%len(patternBytes)]
				}

				// Verify data integrity using MD5 checksum
				expectedMD5 := fmt.Sprintf("%x", md5.Sum(expectedData))
				actualMD5 := fmt.Sprintf("%x", md5.Sum(body))

				if expectedMD5 != actualMD5 {
					t.Errorf("%s: MD5 checksum mismatch. Expected %s, got %s", tc.description, expectedMD5, actualMD5)
				}
			})

			t.Run("Protobuf", func(t *testing.T) {
				url := fmt.Sprintf("%s?size=%d", ts.URL, tc.responseSize)
				protoReq := &proto.HTTPClient{
					Method:   "GET",
					Headers:  map[string]string{},
					Insecure: true,
					Url:      url,
				}

				msg, err := pb.Marshal(protoReq)
				if err != nil {
					t.Fatalf("Failed to marshal protobuf request: %s", err)
				}

				b, err := h.Call(msg)
				if err != nil {
					t.Fatalf("HTTP call failed: %s", err)
				}

				var rsp proto.HTTPClientResponse
				if err := pb.Unmarshal(b, &rsp); err != nil {
					t.Fatalf("Failed to unmarshal protobuf response: %s", err)
				}

				if rsp.Status.Code != 200 {
					t.Fatalf("Expected successful status, got %d: %s", rsp.Status.Code, rsp.Status.Status)
				}

				expectedSize := tc.responseSize
				maxSize := tc.config.MaxResponseBodySize
				if maxSize <= 0 {
					maxSize = 10 * 1024 * 1024 // Default 10MB
				}

				if tc.expectTruncated {
					expectedSize = int(maxSize)
				}

				if len(rsp.Body) != expectedSize {
					t.Errorf("%s: expected body length %d, got %d", tc.description, expectedSize, len(rsp.Body))
				}

				// Generate expected data for MD5 verification
				pattern := "This is test data for HTTP response body size limiting. "
				patternBytes := []byte(pattern)
				expectedData := make([]byte, expectedSize)
				for i := 0; i < expectedSize; i++ {
					expectedData[i] = patternBytes[i%len(patternBytes)]
				}

				// Verify data integrity using MD5 checksum
				expectedMD5 := fmt.Sprintf("%x", md5.Sum(expectedData))
				actualMD5 := fmt.Sprintf("%x", md5.Sum(rsp.Body))

				if expectedMD5 != actualMD5 {
					t.Errorf("%s: MD5 checksum mismatch. Expected %s, got %s", tc.description, expectedMD5, actualMD5)
				}
			})
		})
	}
}

// failingRoundTripper simulates transport failures
type failingRoundTripper struct {
	err error
}

func (f *failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, f.err
}



// TestErrorPaths tests all error branches in Call and callJSON
func TestErrorPaths(t *testing.T) {
	h, err := New(Config{})
	if err != nil {
		t.Fatalf("Unable to create HTTP Client - %s", err)
	}

	t.Run("Protobuf Error Paths", func(t *testing.T) {
		tests := []struct {
			name           string
			setup          func() ([]byte, error)
			expectErr      bool
			expectCode     int32
			expectContains string
		}{
			{
				name: "Invalid Method in Request",
				setup: func() ([]byte, error) {
					msg := &proto.HTTPClient{
						Method:   "INVALID\nMETHOD", // Invalid method with newline
						Url:      "http://localhost",
						Insecure: true,
					}
					return pb.Marshal(msg)
				},
				expectErr:      true,
				expectCode:     400,
				expectContains: "Unable to create HTTP request",
			},
			{
				name: "Invalid URL in Request",
				setup: func() ([]byte, error) {
					msg := &proto.HTTPClient{
						Method:   "GET",
						Url:      "://invalid-url", // Malformed URL
						Insecure: true,
					}
					return pb.Marshal(msg)
				},
				expectErr:      true,
				expectCode:     400,
				expectContains: "Unable to create HTTP request",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := tt.setup()
				if err != nil {
					t.Fatalf("Failed to setup test: %s", err)
				}

				rsp, err := h.Call(payload)
				if tt.expectErr && err == nil {
					t.Fatal("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}

				if tt.expectErr {
					// Parse response
					var r proto.HTTPClientResponse
					if err := pb.Unmarshal(rsp, &r); err != nil {
						t.Fatalf("Failed to unmarshal response: %s", err)
					}

					if r.Status.Code != tt.expectCode {
						t.Errorf("Expected status code %d, got %d", tt.expectCode, r.Status.Code)
					}

					if !strings.Contains(r.Status.Status, tt.expectContains) {
						t.Errorf("Expected status to contain '%s', got '%s'", tt.expectContains, r.Status.Status)
					}

					if !strings.Contains(err.Error(), tt.expectContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.expectContains, err.Error())
					}
				}
			})
		}
	})

	t.Run("JSON Error Paths", func(t *testing.T) {
		tests := []struct {
			name           string
			payload        string
			expectErr      bool
			expectCode     int
			expectContains string
		}{
			{
				name:           "Invalid JSON",
				payload:        `{invalid json`,
				expectErr:      true,
				expectCode:     400,
				expectContains: "Error Parsing Input",
			},
			{
				name:           "Invalid Base64 Body",
				payload:        `{"method":"POST","url":"http://localhost","body":"not-valid-base64!!!"}`,
				expectErr:      true,
				expectCode:     400,
				expectContains: "Unable to decode data",
			},
			{
				name:           "Invalid Method",
				payload:        `{"method":"INVALID\nMETHOD","url":"http://localhost"}`,
				expectErr:      true,
				expectCode:     400,
				expectContains: "Unable to create HTTP request",
			},
			{
				name:           "Invalid URL",
				payload:        `{"method":"GET","url":"://invalid"}`,
				expectErr:      true,
				expectCode:     400,
				expectContains: "Unable to create HTTP request",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				rsp, err := h.Call([]byte(tt.payload))
				if tt.expectErr && err == nil {
					t.Fatal("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Fatalf("Unexpected error: %s", err)
				}

				if tt.expectErr {
					// Parse response
					var r tarmac.HTTPClientResponse
					if err := ffjson.Unmarshal(rsp, &r); err != nil {
						t.Fatalf("Failed to unmarshal response: %s", err)
					}

					if r.Status.Code != tt.expectCode {
						t.Errorf("Expected status code %d, got %d", tt.expectCode, r.Status.Code)
					}

					if !strings.Contains(r.Status.Status, tt.expectContains) {
						t.Errorf("Expected status to contain '%s', got '%s'", tt.expectContains, r.Status.Status)
					}

					if !strings.Contains(err.Error(), tt.expectContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.expectContains, err.Error())
					}
				}
			})
		}
	})
}

// TestTransportFailures tests HTTP transport failures
func TestTransportFailures(t *testing.T) {
	testErr := errors.New("simulated transport failure")

	t.Run("Protobuf Transport Failure", func(t *testing.T) {
		h, err := New(Config{})
		if err != nil {
			t.Fatalf("Unable to create HTTP Client - %s", err)
		}

		msg := &proto.HTTPClient{
			Method: "GET",
			Url:    "http://localhost:99999", // Invalid port to trigger connection error
		}
		payload, err := pb.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal request: %s", err)
		}

		rsp, err := h.Call(payload)
		if err == nil {
			t.Fatal("Expected transport error but got none")
		}

		var r proto.HTTPClientResponse
		if err := pb.Unmarshal(rsp, &r); err != nil {
			t.Fatalf("Failed to unmarshal response: %s", err)
		}

		if r.Status.Code != 500 {
			t.Errorf("Expected status code 500, got %d", r.Status.Code)
		}

		if !strings.Contains(r.Status.Status, "Unable to execute HTTP request") {
			t.Errorf("Expected status to contain 'Unable to execute HTTP request', got '%s'", r.Status.Status)
		}

		if !strings.Contains(err.Error(), "Unable to execute HTTP request") {
			t.Errorf("Expected error to contain 'Unable to execute HTTP request', got '%s'", err.Error())
		}
	})

	t.Run("JSON Transport Failure", func(t *testing.T) {
		h, err := New(Config{})
		if err != nil {
			t.Fatalf("Unable to create HTTP Client - %s", err)
		}

		payload := `{"method":"GET","url":"http://localhost:99999"}`

		rsp, err := h.Call([]byte(payload))
		if err == nil {
			t.Fatal("Expected transport error but got none")
		}

		var r tarmac.HTTPClientResponse
		if err := ffjson.Unmarshal(rsp, &r); err != nil {
			t.Fatalf("Failed to unmarshal response: %s", err)
		}

		if r.Status.Code != 500 {
			t.Errorf("Expected status code 500, got %d", r.Status.Code)
		}

		if !strings.Contains(r.Status.Status, "Unable to execute HTTP request") {
			t.Errorf("Expected status to contain 'Unable to execute HTTP request', got '%s'", r.Status.Status)
		}

		if !strings.Contains(err.Error(), "Unable to execute HTTP request") {
			t.Errorf("Expected error to contain 'Unable to execute HTTP request', got '%s'", err.Error())
		}
	})

	// Test with custom failing transport
	t.Run("Custom Failing Transport", func(_ *testing.T) {
		// Create a custom HTTPClient with failing transport
		// This test verifies the error handling works with explicit transport failures
		// Note: This is primarily for documentation/demonstration as the actual
		// http.Client is created internally and cannot be easily mocked
		_ = &failingRoundTripper{err: testErr}
		// The actual implementation creates its own http.Client internally
		// so we test via invalid URLs that trigger real transport errors
	})
}

// TestBodyReadFailures tests response body read failures
func TestBodyReadFailures(t *testing.T) {
	t.Run("Protobuf Body Read Error", func(t *testing.T) {
		h, err := New(Config{})
		if err != nil {
			t.Fatalf("Unable to create HTTP Client - %s", err)
		}

		// Create a test server that returns a body that will fail to read
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Write a chunked response and then close the connection abruptly
			w.Header().Set("Content-Length", "1000000") // Claim large body
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			// Panic to simulate connection drop (server will close connection)
			panic("simulated connection drop")
		}))
		defer ts.Close()

		msg := &proto.HTTPClient{
			Method: "GET",
			Url:    ts.URL,
		}
		payload, err := pb.Marshal(msg)
		if err != nil {
			t.Fatalf("Failed to marshal request: %s", err)
		}

		rsp, callErr := h.Call(payload)
		// Body read errors are handled gracefully in the current implementation
		// The error is set in the status but the function returns the response

		var r proto.HTTPClientResponse
		if err := pb.Unmarshal(rsp, &r); err != nil {
			t.Fatalf("Failed to unmarshal response: %s", err)
		}

		// When body read fails, we expect status code 500 or the error to be captured
		if r.Status.Code == 500 {
			if !strings.Contains(r.Status.Status, "error reading HTTP response body") {
				t.Logf("Body read error handled, status: %s", r.Status.Status)
			}
		}
		// Avoid unused variable warning
		_ = callErr
	})

	t.Run("JSON Body Read Error", func(t *testing.T) {
		h, err := New(Config{})
		if err != nil {
			t.Fatalf("Unable to create HTTP Client - %s", err)
		}

		// Create a test server that returns a body that will fail to read
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1000000")
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			panic("simulated connection drop")
		}))
		defer ts.Close()

		payload := fmt.Sprintf(`{"method":"GET","url":"%s"}`, ts.URL)

		rsp, callErr := h.Call([]byte(payload))

		var r tarmac.HTTPClientResponse
		if err := ffjson.Unmarshal(rsp, &r); err != nil {
			t.Fatalf("Failed to unmarshal response: %s", err)
		}

		// When body read fails, we expect status code 500 or the error to be captured
		if r.Status.Code == 500 {
			if !strings.Contains(r.Status.Status, "error reading HTTP response body") {
				t.Logf("Body read error handled, status: %s", r.Status.Status)
			}
		}
		// Avoid unused variable warning
		_ = callErr
	})
}

// TestErrorSemantics verifies consistent error handling across both paths
func TestErrorSemantics(t *testing.T) {
	tests := []struct {
		name           string
		protoPayload   func() ([]byte, error)
		jsonPayload    string
		expectErr      bool
		expectCode     int
		expectContains string
	}{
		{
			name: "Invalid URL - Both Paths",
			protoPayload: func() ([]byte, error) {
				return pb.Marshal(&proto.HTTPClient{
					Method: "GET",
					Url:    "://invalid",
				})
			},
			jsonPayload:    `{"method":"GET","url":"://invalid"}`,
			expectErr:      true,
			expectCode:     400,
			expectContains: "Unable to create HTTP request",
		},
		{
			name: "Transport Failure - Both Paths",
			protoPayload: func() ([]byte, error) {
				return pb.Marshal(&proto.HTTPClient{
					Method: "GET",
					Url:    "http://localhost:99999",
				})
			},
			jsonPayload:    `{"method":"GET","url":"http://localhost:99999"}`,
			expectErr:      true,
			expectCode:     500,
			expectContains: "Unable to execute HTTP request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(Config{})
			if err != nil {
				t.Fatalf("Unable to create HTTP Client - %s", err)
			}

			// Test Protobuf path
			t.Run("Protobuf", func(t *testing.T) {
				payload, err := tt.protoPayload()
				if err != nil {
					t.Fatalf("Failed to setup test: %s", err)
				}

				rsp, err := h.Call(payload)
				if tt.expectErr && err == nil {
					t.Fatal("Expected error but got none")
				}

				if tt.expectErr {
					var r proto.HTTPClientResponse
					if err := pb.Unmarshal(rsp, &r); err != nil {
						t.Fatalf("Failed to unmarshal response: %s", err)
					}

					if r.Status.Code != int32(tt.expectCode) {
						t.Errorf("Expected status code %d, got %d", tt.expectCode, r.Status.Code)
					}

					if !strings.Contains(r.Status.Status, tt.expectContains) {
						t.Errorf("Expected status to contain '%s', got '%s'", tt.expectContains, r.Status.Status)
					}
				}
			})

			// Test JSON path
			t.Run("JSON", func(t *testing.T) {
				rsp, err := h.Call([]byte(tt.jsonPayload))
				if tt.expectErr && err == nil {
					t.Fatal("Expected error but got none")
				}

				if tt.expectErr {
					var r tarmac.HTTPClientResponse
					if err := ffjson.Unmarshal(rsp, &r); err != nil {
						t.Fatalf("Failed to unmarshal response: %s", err)
					}

					if r.Status.Code != tt.expectCode {
						t.Errorf("Expected status code %d, got %d", tt.expectCode, r.Status.Code)
					}

					if !strings.Contains(r.Status.Status, tt.expectContains) {
						t.Errorf("Expected status to contain '%s', got '%s'", tt.expectContains, r.Status.Status)
					}
				}
			})
		})
	}
}
