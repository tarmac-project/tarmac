package app

import (
	"encoding/base64"
	"fmt"
	"github.com/madflojo/tarmac"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type httpcallCase struct {
	err      bool
	pass     bool
	httpCode int
	name     string
	call     string
	json     string
}

func TestHTTPCall(t *testing.T) {
	h := &httpcall{}

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
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if strings.ToUpper(string(body)) != r.Method {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Fprintf(w, "%s", body)
		default:
			return
		}

	}))

	var tc []httpcallCase

	// Create a collection of test cases
	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple GET",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"GET","headers":{"teapot": "true"},"insecure":true,"url":"%s"}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      true,
		pass:     false,
		httpCode: 0,
		name:     "Simple GET without SkipVerify",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"GET","headers":{"teapot": "true"},"url":"%s"}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple HEAD",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"HEAD","headers":{"teapot": "true"},"insecure":true,"url":"%s"}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple DELETE",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"DELETE","headers":{"teapot": "true"},"insecure":true,"url":"%s"}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple POST",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"POST","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"UE9TVA=="}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 400,
		name:     "Invalid POST",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"POST","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"NotValid"}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple PUT",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PUT","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"UFVU"}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 400,
		name:     "Invalid PUT",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PUT","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"NotValid"}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 200,
		name:     "Simple PATCH",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PATCH","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"UEFUQ0g="}`, ts.URL),
	})

	tc = append(tc, httpcallCase{
		err:      false,
		pass:     true,
		httpCode: 400,
		name:     "Simple PATCH",
		call:     "Call",
		json:     fmt.Sprintf(`{"method":"PATCH","headers":{"teapot": "true"},"insecure":true,"url":"%s","body":"NotValid"}`, ts.URL),
	})

	// Loop through test cases executing and validating
	for _, c := range tc {
		switch c.call {
		case "Call":
			t.Run(c.name+" Call", func(t *testing.T) {
				// Call http callback
				b, err := h.Call([]byte(c.json))
				if err != nil && !c.err {
					t.Fatalf("HTTPCall Callback failed unexpectedly - %s", err)
				}
				if err == nil && c.err {
					t.Fatalf("HTTPCall Callback unexpectedly passed")
				}

				// Validate Response
				var rsp tarmac.HTTPCallResponse
				err = ffjson.Unmarshal(b, &rsp)
				if err != nil {
					t.Fatalf("HTTPCall Callback Set replied with an invalid JSON - %s", err)
				}

				// Tarmac Response
				if rsp.Status.Code == 200 && !c.pass {
					t.Fatalf("HTTPCall Callback Set returned an unexpected success - %+v", rsp)
				}
				if rsp.Status.Code != 200 && c.pass {
					t.Fatalf("HTTPCall Callback Set returned an unexpected failure - %+v", rsp)
				}

				// HTTP Response
				if rsp.Code != c.httpCode {
					t.Fatalf("HTTPCall returned an unexpected response code - %+v", rsp)
					return
				}

				// Validate Response Header
				v, ok := rsp.Headers["server"]
				if rsp.Code == 200 && (!ok || v != "tarmac") {
					t.Errorf("HTTPCall returned an unexpected header - %+v", rsp)
				}

				// Validate Payload
				if len(rsp.Body) > 0 {
					body, err := base64.StdEncoding.DecodeString(rsp.Body)
					if err != nil {
						t.Fatalf("Error decoding HTTPCall returned body - %s", err)
					}
					switch fmt.Sprintf("%s", body) {
					case "PUT", "POST", "PATCH":
						return
					default:
						t.Errorf("HTTPCall returned unexpected payload - %s", body)
					}
				}
			})
		}
	}
}
