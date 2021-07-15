package app

import (
	"encoding/base64"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/madflojo/tarmac"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

// server is used as an interface for managing the HTTP server.
type server struct {
	// httpServer is the primary HTTP server.
	httpServer *http.Server

	// httpRouter is used to store and access the HTTP Request Router.
	httpRouter *httprouter.Router

	// kvStore provides access to the key:value datastore callback functions for WASM Modules.
	kvStore *kvStore

	// logger provides access to the logging callback functions for WASM Modules.
	logger *logger
}

// Health is used to handle HTTP Health requests to this service. Use this for liveness
// probes or any other checks which only validate if the services is running.
func (s *server) Health(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}

// Ready is used to handle HTTP Ready requests to this service. Use this for readiness
// probes or any checks that validate the service is ready to accept traffic.
func (s *server) Ready(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Check other stuff here like KV connectivity, health of dependent services, etc.
	if cfg.GetBool("enable_kvstore") {
		err := kv.HealthCheck()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// middleware is used to intercept incoming HTTP calls and apply general functions upon
// them. e.g. Metrics, Logging...
func (s *server) middleware(n httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Set the Tarmac server response header
		w.Header().Set("Server", "tarmac")

		// Log the basics
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("HTTP Request to %s", r.URL)

		// Call registered handler
		n(w, r, ps)
	}
}

// WASMHandler is the primary HTTP handler for WASM Module traffic. This handler will load the
// specified module and create an execution environment for that module.
func (s *server) WASMHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Read the HTTP Payload
	var payload []byte
	var err error
	if r.Method == "POST" || r.Method == "PUT" {
		payload, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.WithFields(logrus.Fields{
				"method":         r.Method,
				"remote-addr":    r.RemoteAddr,
				"http-protocol":  r.Proto,
				"headers":        r.Header,
				"content-length": r.ContentLength,
			}).Debugf("Error reading HTTP payload - %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Create Request type
	req := tarmac.ServerRequest{
		Headers: map[string]string{
			"request_type": "http",
			"http_method":  r.Method,
			"http_path":    r.URL.Path,
			"remote_addr":  r.RemoteAddr,
		},
		Payload: base64.StdEncoding.EncodeToString(payload),
	}

	// Append Request Headers
	for k := range r.Header {
		req.Headers[strings.ToLower(k)] = r.Header.Get(k)
	}

	// Convert request to JSON payload
	reqData, err := ffjson.Marshal(req)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error creating request type for WASM module - %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Fetch Module and run with payload
	m, err := engine.Module("default")
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error loading wasi environment - %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Execute the WASM HTTP Handler
	var rsp tarmac.ServerResponse
	rspData, err := m.Run("http:"+r.Method, reqData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error executing function - %s", err)
	}

	// Unmarshal WASM JSON response
	err = ffjson.Unmarshal(rspData, &rsp)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error parsing response type from WASM module - %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Decode Response Payload
	rspPayload, err := base64.StdEncoding.DecodeString(rsp.Payload)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error decoing base64 payload response from WASM module - %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If Response indicates error print it out
	if rsp.Status.Code > 399 {
		// Return status code and print stdout
		w.WriteHeader(rsp.Status.Code)
		fmt.Fprintf(w, "%s", rsp.Status.Status)
		return
	}

	// Assume if no status code everything worked as expected
	if rsp.Status.Code == 0 {
		rsp.Status.Code = 200
	}

	// Return status code and print stdout
	w.WriteHeader(rsp.Status.Code)
	fmt.Fprintf(w, "%s", rspPayload)
}
