package sdk

import (
	"encoding/base64"
	"fmt"
	"github.com/valyala/fastjson"
	"strings"
)

// HTTPClient provides an interface to make outbound HTTP calls.
type HTTPClient struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// newHTTPClient returns an HTTPClient for making outbound HTTP calls.
func newHTTPClient(cfg Config) *HTTPClient {
	return &HTTPClient{namespace: cfg.Namespace, hostCall: cfg.hostCall}
}

// HTTPResponse is returned from successful client calls.
type HTTPResponse struct {
	// StatusCode is the HTTP status code returned by the Client.
	StatusCode int

	// Headers are the returned HTTP headers from the server response.
	Headers map[string]string

	// Body is the returned HTTP body from the server response.
	Body []byte
}

// Get will perform a GET request using the URL specified.
func (h *HTTPClient) Get(url string) (HTTPResponse, error) {
	return h.Do("GET", nil, url, false, nil)
}

// Delete will perform a DELETE request using the URL specified.
func (h *HTTPClient) Delete(url string) (HTTPResponse, error) {
	return h.Do("DELETE", nil, url, false, nil)
}

// Post will perform a POST request using the URL and Payload specified.
func (h *HTTPClient) Post(url string, payload []byte) (HTTPResponse, error) {
	return h.Do("POST", nil, url, false, payload)
}

// Put will perform a PUT request using the URL and Payload specified.
func (h *HTTPClient) Put(url string, payload []byte) (HTTPResponse, error) {
	return h.Do("PUT", nil, url, false, payload)
}

// Do will perform HTTP requests using the specified parameters.
// Valid Methods are GET, POST, PUT, and DELETE.
func (h *HTTPClient) Do(method string, headers map[string]string, url string, insecure bool, payload []byte) (HTTPResponse, error) {
	// Validate user provided method
	if method != "GET" && method != "POST" && method != "DELETE" && method != "PUT" {
		return HTTPResponse{}, fmt.Errorf("invalid method specified")
	}

	// Validate URL is not empty
	if url == "" {
		return HTTPResponse{}, fmt.Errorf("url cannot be empty")
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
		return HTTPResponse{}, fmt.Errorf("unable to call HTTPClient - %s", err)
	}

	// Parse response JSON
	v, err := fastjson.ParseBytes(b)
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("unable to parse HTTPClient resposne - %s", err)
	}

	rsp := HTTPResponse{}

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
		return rsp, fmt.Errorf("unable to decode HTTPClient response - %s", err)
	}

	return rsp, nil
}
