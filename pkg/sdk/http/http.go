/*
Package http is a client package for WASM functions running within a Tarmac server.

This package provides user-friendly HTTP client functions that can interact with external HTTP services. Guest WASM functions running inside Tarmac can import and call this HTTP client.
*/
package http

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/valyala/fastjson"
)

// Client provides an interface to make outbound HTTP calls.
type Client struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// Config provides users with the ability to specify namespaces, function handlers and other key information required to execute the
// function.
type Config struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	Namespace string

	// HostCall is used internally for host callbacks. This is mainly here for testing.
	HostCall func(string, string, string, []byte) ([]byte, error)
}

// New creates a new Client with the provided configuration.
func New(cfg Config) (*Client, error) {
	// Set default namespace
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// Verify HostCall is set
	if cfg.HostCall == nil {
		return &Client{}, fmt.Errorf("HostCall cannot be nil")
	}

	return &Client{namespace: cfg.Namespace, hostCall: cfg.HostCall}, nil
}

// Response is returned from successful client calls.
type Response struct {
	// StatusCode is the HTTP status code returned by the Client.
	StatusCode int

	// Headers are the returned HTTP headers from the server response.
	Headers map[string]string

	// Body is the returned HTTP body from the server response.
	Body []byte
}

// Get will perform a GET request using the URL specified.
func (h *Client) Get(url string) (Response, error) {
	return h.Do("GET", nil, url, false, nil)
}

// Delete will perform a DELETE request using the URL specified.
func (h *Client) Delete(url string) (Response, error) {
	return h.Do("DELETE", nil, url, false, nil)
}

// Post will perform a POST request using the URL and Payload specified.
func (h *Client) Post(url string, payload []byte) (Response, error) {
	return h.Do("POST", nil, url, false, payload)
}

// Put will perform a PUT request using the URL and Payload specified.
func (h *Client) Put(url string, payload []byte) (Response, error) {
	return h.Do("PUT", nil, url, false, payload)
}

// Do will perform HTTP requests using the specified parameters.
// Valid Methods are GET, POST, PUT, and DELETE.
func (h *Client) Do(method string, headers map[string]string, url string, insecure bool, payload []byte) (Response, error) {
	// Validate user provided method
	if method != "GET" && method != "POST" && method != "DELETE" && method != "PUT" {
		return Response{}, fmt.Errorf("invalid method specified")
	}

	// Validate URL is not empty
	if url == "" {
		return Response{}, fmt.Errorf("url cannot be empty")
	}

	// Build headers string
	hh := []string{}
	for k, v := range headers {
		hh = append(hh, fmt.Sprintf(`"%s":"%s"`, k, v))
	}

	// Encode HTTP Payload
	d := ""
	if payload != nil {
		d = base64.StdEncoding.EncodeToString(payload)
	}

	// Build Callback JSON
	r := fmt.Sprintf(`{"method":"%s", "headers": {%s},"url":"%s","body":"%s", "insecure": %t}`, method, strings.Join(hh, ", "), url, d, insecure)

	// Perform Host Callback
	b, err := h.hostCall(h.namespace, "httpclient", "call", []byte(r))
	if err != nil {
		return Response{}, fmt.Errorf("unable to call Client - %s", err)
	}

	// Parse response JSON
	v, err := fastjson.ParseBytes(b)
	if err != nil {
		return Response{}, fmt.Errorf("unable to parse Client resposne - %s", err)
	}

	rsp := Response{}

	// Extract Status Code
	rsp.StatusCode = v.GetInt("code")

	// Extract Headers
	rsp.Headers = make(map[string]string)
	vMap := v.GetObject("headers")
	if vMap != nil {
		vMap.Visit(func(k []byte, val *fastjson.Value) {
			rsp.Headers[string(k)] = string(val.GetStringBytes())
		})
	}

	// Extract and Decode Payload
	rsp.Body, err = base64.StdEncoding.DecodeString(string(v.GetStringBytes("body")))
	if err != nil {
		return rsp, fmt.Errorf("unable to decode Client response - %s", err)
	}

	return rsp, nil
}
