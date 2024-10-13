/*
Package httpclient is part of the Tarmac suite of Host Callback packages. This package provides users with the
ability to provide WASM functions with a host callback interface that provides HTTP client capabilities.

	import (
		"github.com/tarmac-project/tarmac/pkg/callbacks/httpclient"
	)

	func main() {
		// Create instance of httpclient to register for callback execution
		httpclient := httpclient.New(httpclient.Config{})

		// Create Callback router and register httpclient
		router := callbacks.New()
		router.RegisterCallback("httpclient", "Call", httpclient.Call)
	}
*/
package httpclient

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/tarmac-project/tarmac"

	"github.com/tarmac-project/protobuf-go/sdk"
	proto "github.com/tarmac-project/protobuf-go/sdk/http"
	pb "google.golang.org/protobuf/proto"
)

// HTTPClient provides access to Host Callbacks that interact with an HTTP client. These callbacks offer all of the logic
// and error handlings of interacting with an HTTP server. Users will send the specified JSON request and receive
// an appropriate JSON response.
type HTTPClient struct{}

// Config is provided to users to configure the Host Callback. All Tarmac Callbacks follow the same configuration
// format; each Config struct gives the specific Host Callback unique functionality.
type Config struct{}

// New will create and return a new HTTPClient instance that users can register as a Tarmac Host Callback function.
// Users can provide any custom HTTP Client configurations using the configuration options supplied.
func New(_ Config) (*HTTPClient, error) {
	hc := &HTTPClient{}
	return hc, nil
}

// Call will perform the desired HTTP request using the supplied JSON as configuration. Logging, error handling, and
// base64 decoding of payload data are all handled via this function. Note, this function expects the
// HTTPClientRequest JSON type as input and will return a KVStoreGetResponse JSON.
func (hc *HTTPClient) Call(b []byte) ([]byte, error) {
	// Parse incoming Request
	msg := &proto.HTTPClient{}
	err := pb.Unmarshal(b, msg)
	if err != nil {
		// Assume JSON for backwards compatibility
		return hc.callJSON(b)
	}

	// Create HTTPClientResponse
	r := &proto.HTTPClientResponse{}
	r.Status = &sdk.Status{Code: 200, Status: "OK"}

	// Create HTTP Client
	var request *http.Request
	var c *http.Client
	tr := &http.Transport{}
	if msg.Insecure {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	c = &http.Client{Transport: tr}

	// Create HTTP Request
	request, err = http.NewRequest(msg.Method, msg.Url, bytes.NewBuffer(msg.Body))
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = fmt.Sprintf("Unable to create HTTP request - %s", err)
		// Marshal a response to return to caller
		rsp, err := pb.Marshal(r)
		if err != nil {
			return []byte(""), fmt.Errorf("unable to marshal HTTPClient:call response")
		}
		return rsp, fmt.Errorf("%s", r.Status.Status)
	}

	// Set user-supplied headers
	for k, v := range msg.Headers {
		request.Header.Set(k, v)
	}

	// Execute HTTP Call
	response, err := c.Do(request)
	if err != nil {
		r.Status.Code = 500
		r.Status.Status = fmt.Sprintf("Unable to execute HTTP request - %s", err)
		// Marshal a response to return to caller
		rsp, err := pb.Marshal(r)
		if err != nil {
			return []byte(""), fmt.Errorf("unable to marshal HTTPClient:call response")
		}
		return rsp, fmt.Errorf("%s", r.Status.Status)
	}

	// Populate Response with Response
	if response != nil { // nolint
		defer response.Body.Close()
		r.Code = int32(response.StatusCode)
		r.Headers = make(map[string]string)
		for k := range response.Header {
			r.Headers[strings.ToLower(k)] = response.Header.Get(k)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unexpected error reading HTTP response body - %s", err)
		}
		r.Body = body
	}

	// Marshal a response to return to caller
	rsp, err := pb.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal HTTPClient:call response")
	}

	// Return response to caller
	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}

func (hc *HTTPClient) callJSON(b []byte) ([]byte, error) {
	// Start Response Message assuming everything is good
	r := tarmac.HTTPClientResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	// Parse incoming Request
	var rq tarmac.HTTPClient
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = "Error Parsing Input"
	}

	// Decode data to store
	data, err := base64.StdEncoding.DecodeString(rq.Body)
	if err != nil {
		r.Status.Code = 400
		r.Status.Status = fmt.Sprintf("Unable to decode data - %s", err)
	}

	// Create HTTP Client
	var request *http.Request
	var c *http.Client
	if r.Status.Code == 200 {
		tr := &http.Transport{}
		if rq.Insecure {
			tr.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		c = &http.Client{Transport: tr}

		// Create HTTP Request
		request, err = http.NewRequest(rq.Method, rq.URL, bytes.NewBuffer(data))
		if err != nil {
			r.Status.Code = 400
			r.Status.Status = fmt.Sprintf("Unable to create HTTP request - %s", err)
		}

		// Set user-supplied headers
		for k, v := range rq.Headers {
			request.Header.Set(k, v)
		}
	}

	// Execute HTTP Call
	if r.Status.Code == 200 {
		response, err := c.Do(request)
		if err != nil {
			r.Status.Code = 500
			r.Status.Status = fmt.Sprintf("Unable to execute HTTP request - %s", err)
		}

		// Populate Response with Response
		if response != nil { // nolint
			defer response.Body.Close()
			r.Code = response.StatusCode
			r.Headers = make(map[string]string)
			for k := range response.Header {
				r.Headers[strings.ToLower(k)] = response.Header.Get(k)
			}
			body, err := io.ReadAll(response.Body)
			if err != nil {
				r.Status.Code = 500
				r.Status.Status = fmt.Sprintf("Unexpected error reading HTTP response body - %s", err)
			}
			r.Body = base64.StdEncoding.EncodeToString(body)
		}
	}

	// Marshal a response JSON to return to caller
	rsp, err := ffjson.Marshal(r)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to marshal HTTPClient:call response")
	}

	// Return response to caller
	if r.Status.Code == 200 {
		return rsp, nil
	}
	return rsp, fmt.Errorf("%s", r.Status.Status)
}
