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
	"regexp"
	"strings"
	"time"
)

// isPProf is a regex that validates if the given path is used for PProf
var isPProf = regexp.MustCompile(`.*debug\/pprof.*`)

// server is used as an interface for managing the HTTP server.
type server struct {
	// httpServer is the primary HTTP server.
	httpServer *http.Server

	// httpRouter is used to store and access the HTTP Request Router.
	httpRouter *httprouter.Router
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
		now := time.Now()

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

		// Verify if PProf
		if isPProf.MatchString(r.URL.Path) && !cfg.GetBool("enable_pprof") {
			log.WithFields(logrus.Fields{
				"method":         r.Method,
				"remote-addr":    r.RemoteAddr,
				"http-protocol":  r.Proto,
				"headers":        r.Header,
				"content-length": r.ContentLength,
			}).Debugf("Request to PProf Address failed, PProf disabled")
			w.WriteHeader(http.StatusForbidden)

			stats.srv.WithLabelValues(r.URL.Path).Observe(time.Since(now).Seconds())
			return
		}

		// Call registered handler
		n(w, r, ps)
		stats.srv.WithLabelValues(r.URL.Path).Observe(time.Since(now).Seconds())
	}
}

// handlerWrapper is used to wrap http.Handler functions with the server middleware.
func (s *server) handlerWrapper(h http.Handler) httprouter.Handle {
	return s.middleware(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.ServeHTTP(w, r)
	})
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

	// Execute WASM Module
	rsp, err := runWASM("default", "http:"+r.Method, req)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error executing WASM module - %s", err)
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

// runWASM will load and execute the specified WASM module.
func runWASM(module, handler string, rq tarmac.ServerRequest) (tarmac.ServerResponse, error) {
	var rsp tarmac.ServerResponse
	now := time.Now()

	// Convert request to JSON payload
	d, err := ffjson.Marshal(rq)
	if err != nil {
		stats.wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
		return rsp, fmt.Errorf("unable to marshal server request - %s", err)
	}

	// Fetch Module and run with payload
	m, err := engine.Module(module)
	if err != nil {
		stats.wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
		return rsp, fmt.Errorf("unable to load wasi environment - %s", err)
	}

	// Execute the WASM Handler
	data, err := m.Run(handler, d)
	if err != nil {
		stats.wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
		return rsp, fmt.Errorf("failed to execute wasm module - %s", err)
	}

	// Unmarshal WASM JSON response
	err = ffjson.Unmarshal(data, &rsp)
	if err != nil {
		stats.wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
		return rsp, fmt.Errorf("failed to unmarshal response - %s - %s", err, data)
	}

	stats.wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
	return rsp, nil
}
