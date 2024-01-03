package http

import (
	"fmt"
	"testing"

	"github.com/pquerna/ffjson/ffjson"
)

type HTTPDoTestCase struct {
	name     string
	err      bool
	hostCall func(string, string, string, []byte) ([]byte, error)
	method   string
	url      string
	headers  map[string]string
	insecure bool
	payload  []byte
}

func TestHTTPDo(t *testing.T) {
	var tt []HTTPDoTestCase

	// HTTP Do Post
	tc := HTTPDoTestCase{
		name:     "Valid HTTP Post",
		err:      false,
		method:   "POST",
		url:      "http://example.com",
		insecure: false,
		headers:  make(map[string]string),
		payload:  []byte("Testing 1 2 3"),
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			if namespace != "default" || capability != "httpclient" || function != "call" {
				t.Errorf("Incorrect arguments to hostCall - %s, %s, %s", namespace, capability, function)
			}

			var req map[string]interface{}
			err := ffjson.Unmarshal(input, &req)
			if err != nil {
				t.Errorf("Unexpected error parsing input JSON: %v", err)
			}

			if req["method"] != "POST" {
				t.Errorf("unexpected method value: %v", req)
			}

			// Validate method
			method, ok := req["method"].(string)
			if !ok || method != "POST" {
				t.Errorf("Invalid or missing method value: %v", req)
			}

			// Validate URL
			_, ok = req["url"].(string)
			if !ok {
				t.Errorf("Invalid or missing url value: %v", req)
			}

			// Validate insecure
			insecure, ok := req["insecure"].(bool)
			if !ok {
				t.Errorf("Invalid or missing insecure value: %v", req)
			}
			if insecure {
				t.Errorf("Invalid insecure value")
			}

			// Validate headers
			headers, ok := req["headers"].(map[string]interface{})
			if !ok {
				t.Errorf("Invalid or missing headers value: %v", req)
			}
			x, ok := headers["testing"]
			if !ok {
				t.Errorf("Missing testing header: %v", headers)
			}
			if x != "testing" {
				t.Errorf("Testing header is invalid")
			}

			// Validate payload
			_, ok = req["body"].(string)
			if !ok {
				t.Errorf("Invalid or missing body value: %v", req)
			}

			return []byte(fmt.Sprintf(`{"code": 200,"headers":{"testing":"testing"},"body":"%s","status":{"code":200,"status":"OK"}}`, req["body"])), nil

		},
	}
	tc.headers["testing"] = "testing"
	tt = append(tt, tc)

	tc = HTTPDoTestCase{
		name:     "Valid HTTP GET",
		err:      false,
		method:   "GET",
		url:      "http://example.com",
		insecure: false,
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			// validate hostCall arguments
			if namespace != "default" || capability != "httpclient" || function != "call" {
				t.Errorf("Incorrect arguments to hostCall - %s, %s, %s", namespace, capability, function)
			}

			return []byte(`{"code": 200,"headers":{},"body":"dGVzdA==","status":{"code":200,"status":"OK"}}`), nil
		},
	}
	tt = append(tt, tc)

	tc = HTTPDoTestCase{
		name:     "HTTP request with empty payload",
		err:      false,
		method:   "POST",
		url:      "http://example.com",
		insecure: false,
		headers:  make(map[string]string),
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			// validate hostCall arguments and input
			if namespace != "default" || capability != "httpclient" || function != "call" {
				t.Errorf("Incorrect arguments to hostCall - %s, %s, %s", namespace, capability, function)
			}

			return []byte(`{"code": 200,"headers":{},"body":"","status":{"code":200,"status":"OK"}}`), nil
		},
	}
	tt = append(tt, tc)

	tc = HTTPDoTestCase{
		name:     "HTTP request with invalid response payload",
		err:      true,
		method:   "GET",
		url:      "http://example.com",
		insecure: false,
		headers:  make(map[string]string),
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			return []byte(`{"code": 200,"headers":{},"body":"THIS IS NOT BASE64","status":{"code":200,"status":"OK"}}`), nil
		},
	}
	tt = append(tt, tc)

	tc = HTTPDoTestCase{
		name:     "HTTP request with invalid URL",
		err:      true,
		method:   "POST",
		url:      "",
		insecure: false,
		headers:  make(map[string]string),
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			// this should not be called since the URL is invalid
			t.Errorf("hostCall should not have been called")
			return nil, nil
		},
	}
	tc.headers["testing"] = "testing"
	tt = append(tt, tc)

	// HTTP Do Delete
	tc = HTTPDoTestCase{
		name:     "Valid HTTP Delete",
		err:      false,
		method:   "DELETE",
		url:      "http://example.com",
		insecure: false,
		headers:  make(map[string]string),
		payload:  []byte(""),
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			if namespace != "default" || capability != "httpclient" || function != "call" {
				t.Errorf("Incorrect arguments to hostCall - %s, %s, %s", namespace, capability, function)
			}

			var req map[string]interface{}
			err := ffjson.Unmarshal(input, &req)
			if err != nil {
				t.Errorf("Unexpected error parsing input JSON: %v", err)
			}

			if req["method"] != "DELETE" {
				t.Errorf("unexpected method value: %v", req)
			}

			return []byte(fmt.Sprintf(`{"code": 200,"headers":{"testing":"testing"},"body":"%s","status":{"code":200,"status":"OK"}}`, req["body"])), nil

		},
	}
	tc.headers["testing"] = "testing"
	tt = append(tt, tc)

	// HTTP Do Put
	tc = HTTPDoTestCase{
		name:     "Valid HTTP Put",
		err:      false,
		method:   "PUT",
		url:      "http://example.com",
		insecure: false,
		headers:  make(map[string]string),
		payload:  []byte("Testing 1 2 3"),
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			if namespace != "default" || capability != "httpclient" || function != "call" {
				t.Errorf("Incorrect arguments to hostCall - %s, %s, %s", namespace, capability, function)
			}

			var req map[string]interface{}
			err := ffjson.Unmarshal(input, &req)
			if err != nil {
				t.Errorf("Unexpected error parsing input JSON: %v", err)
			}

			if req["method"] != "PUT" {
				t.Errorf("unexpected method value: %v", req)
			}

			// Validate URL
			v, ok := req["url"].(string)
			if !ok {
				t.Errorf("Invalid or missing url value: %v", req)
			}
			if v != "http://example.com" {
				t.Errorf("Invalid http URL")
			}

			// Validate insecure
			insecure, ok := req["insecure"].(bool)
			if !ok {
				t.Errorf("Invalid or missing insecure value: %v", req)
			}
			if insecure {
				t.Errorf("Invalid insecure value")
			}

			// Validate headers
			headers, ok := req["headers"].(map[string]interface{})
			if !ok {
				t.Errorf("Invalid or missing headers value: %v", req)
			}
			x, ok := headers["testing"]
			if !ok {
				t.Errorf("Missing testing header: %v", headers)
			}
			if x != "testing" {
				t.Errorf("Testing header is invalid")
			}

			// Validate payload
			_, ok = req["body"].(string)
			if !ok {
				t.Errorf("Invalid or missing body value: %v", req)
			}

			return []byte(fmt.Sprintf(`{"code": 200,"headers":{"testing":"testing"},"body":"%s","status":{"code":200,"status":"OK"}}`, req["body"])), nil

		},
	}
	tc.headers["testing"] = "testing"
	tt = append(tt, tc)

	tc = HTTPDoTestCase{
		name:     "HTTP request with invalid Method",
		err:      true,
		method:   "NOPE",
		url:      "",
		insecure: false,
		headers:  make(map[string]string),
		hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
			// this should not be called since the URL is invalid
			t.Errorf("hostCall should not have been called")
			return nil, nil
		},
	}
	tc.headers["testing"] = "testing"
	tt = append(tt, tc)

	for _, c := range tt {
		t.Run(c.name, func(t *testing.T) {
			// Create new client
			client, err := New(Config{Namespace: "default", HostCall: c.hostCall})
			if err != nil {
				t.Errorf("unexpected error initiating HTTP client - %s", err)
			}

			// Execute Do
			rsp, err := client.Do(c.method, c.headers, c.url, c.insecure, c.payload)
			if err != nil && c.err {
				return
			}
			if err != nil && !c.err {
				t.Errorf("unexpected error returned - %s", err)
			}
			if err == nil && c.err {
				t.Errorf("expected error got nil - %s", err)
			}

			// Validate response code
			if rsp.StatusCode != 200 {
				t.Errorf("Unexpected response code: %d", rsp.StatusCode)
			}

			// Validate response headers
			if len(c.headers) > 0 {
				if len(rsp.Headers) != 1 || rsp.Headers["testing"] != c.headers["testing"] {
					t.Errorf("Unexpected response headers: %v", rsp.Headers)
				}
			}

			// Validate response body
			if len(c.payload) > 0 {
				if string(rsp.Body) != string(c.payload) {
					t.Errorf("Unexpected response body: %s", rsp.Body)
				}
			}

		})
	}
}

func TestHTTPClientGet(t *testing.T) {
	hc, err := New(Config{Namespace: "default", HostCall: func(string, string, string, []byte) ([]byte, error) {
		return []byte(`{"code": 200,"headers":{},"body":"dGVzdA==","status":{"code":200,"status":"OK"}}`), nil
	}})
	if err != nil {
		t.Errorf("unexpected error initiating HTTP client - %s", err)
	}

	// Test successful GET request
	response, err := hc.Get("http://example.com/get")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response.StatusCode != 200 {
		t.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	// Test unsuccessful GET request
	_, err = hc.Get("")
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}

func TestHTTPClientDelete(t *testing.T) {
	hc, err := New(Config{Namespace: "default", HostCall: func(string, string, string, []byte) ([]byte, error) {
		return []byte(`{"code": 200,"headers":{},"body":"dGVzdA==","status":{"code":200,"status":"OK"}}`), nil
	}})
	if err != nil {
		t.Errorf("unexpected error initiating HTTP client - %s", err)
	}

	// Test successful DELETE request
	response, err := hc.Delete("http://example.com/delete")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response.StatusCode != 200 {
		t.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	// Test unsuccessful DELETE request
	_, err = hc.Delete("")
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}

func TestHTTPClientPost(t *testing.T) {
	hc, err := New(Config{Namespace: "default", HostCall: func(string, string, string, []byte) ([]byte, error) {
		return []byte(`{"code": 200,"headers":{},"body":"dGVzdA==","status":{"code":200,"status":"OK"}}`), nil
	}})
	if err != nil {
		t.Errorf("unexpected error initiating HTTP client - %s", err)
	}

	// Test successful POST request
	response, err := hc.Post("http://example.com/post", []byte(`{"data": "example"}`))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response.StatusCode != 200 {
		t.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	// Test unsuccessful POST request
	_, err = hc.Post("", []byte(`{"data": "example"}`))
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}

func TestHTTPClientPut(t *testing.T) {
	hc, err := New(Config{Namespace: "default", HostCall: func(string, string, string, []byte) ([]byte, error) {
		return []byte(`{"code": 200,"headers":{},"body":"dGVzdA==","status":{"code":200,"status":"OK"}}`), nil
	}})
	if err != nil {
		t.Errorf("unexpected error initiating HTTP client - %s", err)
	}

	// Test successful PUT request
	response, err := hc.Put("http://example.com", []byte(`{"data": "example"}`))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response.StatusCode != 200 {
		t.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	// Test unsuccessful PUT request
	_, err = hc.Put("", []byte(`{"data": "example"}`))
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
}
