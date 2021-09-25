package app

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/madflojo/tarmac"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// httpcall provides access to Host Callbacks that interact with an HTTP client. These callbacks offer all of the logic
// and error handlings of interacting with an HTTP server. Users will send the specified JSON request and receive
// an appropriate JSON response.
type httpcall struct{}

// Call will perform the desired HTTP request using the supplied JSON as configuration. Logging, error handling, and
// base64 decoding of payload data are all handled via this function. Note, this function expects the
// HTTPCallRequest JSON type as input and will return a KVStoreGetResponse JSON.
func (hc *httpcall) Call(b []byte) ([]byte, error) {
	now := time.Now()

	// Start Response Message assuming everything is good
	r := tarmac.HTTPCallResponse{}
	r.Status.Code = 200
	r.Status.Status = "OK"

	// Parse incoming Request
	var rq tarmac.HTTPCall
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
		headers := make(map[string][]string)
		for k, v := range rq.Headers {
			headers[k] = []string{v}
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
		if response != nil {
			r.Code = response.StatusCode
			r.Headers = make(map[string]string)
			for k := range response.Header {
				r.Headers[strings.ToLower(k)] = response.Header.Get(k)
			}
			body, err := ioutil.ReadAll(response.Body)
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
		log.Errorf("Unable to marshal httpcall:call response - %s", err)
		stats.httpcall.WithLabelValues("call").Observe(time.Since(now).Seconds())
		return []byte(""), fmt.Errorf("unable to marshal httpcall:call response")
	}

	// Return response to caller
	if r.Status.Code == 200 {
		stats.httpcall.WithLabelValues("call").Observe(time.Since(now).Seconds())
		return rsp, nil
	}
	stats.httpcall.WithLabelValues("call").Observe(time.Since(now).Seconds())
	return rsp, fmt.Errorf("%s", r.Status.Status)
}
