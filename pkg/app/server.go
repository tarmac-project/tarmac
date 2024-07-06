package app

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/tarmac-project/tarmac/pkg/config"
	"github.com/tarmac-project/tarmac/pkg/sanitize"
)

// isPProf is a regex that validates if the given path is used for PProf
var isPProf = regexp.MustCompile(`.*debug\/pprof.*`)

// Health is used to handle HTTP Health requests to this service. Use this for liveness
// probes or any other checks which only validate if the services is running.
func (srv *Server) Health(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}

// Ready is used to handle HTTP Ready requests to this service. Use this for readiness
// probes or any checks that validate the service is ready to accept traffic.
func (srv *Server) Ready(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	// Check if maintence mode is enabled and return 503
	if srv.cfg.GetBool("enable_maintenance_mode") {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Check other stuff here like KV connectivity, health of dependent services, etc.
	if srv.cfg.GetBool("enable_kvstore") {
		err := srv.kv.HealthCheck()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// middleware is used to intercept incoming HTTP calls and apply general functions upon
// them. e.g. Metrics, Logging...
func (srv *Server) middleware(n httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		now := time.Now()

		// Set the Tarmac server response header
		w.Header().Set("Server", "tarmac")

		// Log the basics
		srv.log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"content-length": r.ContentLength,
		}).Debugf("HTTP Request to %s received", sanitize.String(r.URL.EscapedPath()))

		// Verify if PProf
		if isPProf.MatchString(r.URL.EscapedPath()) && !srv.cfg.GetBool("enable_pprof") {
			srv.log.WithFields(logrus.Fields{
				"method":         r.Method,
				"remote-addr":    r.RemoteAddr,
				"http-protocol":  r.Proto,
				"content-length": r.ContentLength,
				"duration":       time.Since(now).Milliseconds(),
			}).Debugf("Request to PProf Address failed, PProf disabled")
			w.WriteHeader(http.StatusForbidden)

			srv.stats.Srv.WithLabelValues(r.URL.EscapedPath()).Observe(float64(time.Since(now).Milliseconds()))
			return
		}

		// Call registered handler
		n(w, r, ps)
		srv.stats.Srv.WithLabelValues(r.URL.EscapedPath()).Observe(float64(time.Since(now).Milliseconds()))
		srv.log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"content-length": r.ContentLength,
			"duration":       time.Since(now).Milliseconds(),
		}).Debugf("HTTP Request to %s complete", sanitize.String(r.URL.EscapedPath()))
	}
}

// handlerWrapper is used to wrap http.Handler functions with the server middleware.
func (srv *Server) handlerWrapper(h http.Handler) httprouter.Handle {
	return srv.middleware(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.ServeHTTP(w, r)
	})
}

// WASMHandler is the primary HTTP handler for WASM Module traffic. This handler will load the
// specified module and create an execution environment for that module.
func (srv *Server) WASMHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Find Function
	function, err := srv.funcCfg.RouteLookup(fmt.Sprintf("http:%s:%s", r.Method, r.URL.EscapedPath()))
	if err == config.ErrRouteNotFound {
		function = "default"
	}

	// Read the HTTP Payload
	var payload []byte
	if r.Method == "POST" || r.Method == "PUT" {
		payload, err = io.ReadAll(r.Body)
		if err != nil {
			srv.log.WithFields(logrus.Fields{
				"method":         r.Method,
				"remote-addr":    r.RemoteAddr,
				"http-protocol":  r.Proto,
				"content-length": r.ContentLength,
			}).Debugf("Error reading HTTP payload - %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Execute WASM Module
	rsp, err := srv.runWASM(function, "handler", payload)
	if err != nil {
		srv.log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"content-length": r.ContentLength,
		}).Debugf("Error executing WASM module - %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return status code and print stdout
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", rsp)
}

// runWASM will load and execute the specified WASM module.
func (srv *Server) runWASM(module, handler string, rq []byte) ([]byte, error) {
	now := time.Now()

	// Fetch Module and run with payload
	m, err := srv.engine.Module(module)
	if err != nil {
		srv.stats.Wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(float64(time.Since(now).Milliseconds()))
		return []byte(""), fmt.Errorf("unable to load wasi environment - %s", err)
	}

	// Execute the WASM Handler
	rsp, err := m.Run(handler, rq)
	if err != nil {
		srv.stats.Wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(float64(time.Since(now).Milliseconds()))
		return rsp, fmt.Errorf("failed to execute wasm module - %s", err)
	}

	// Return results
	srv.stats.Wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(float64(time.Since(now).Milliseconds()))
	srv.log.WithFields(logrus.Fields{
		"module":   module,
		"handler":  handler,
		"duration": time.Since(now).Milliseconds(),
	}).Debugf("WASM Module Executed")
	return rsp, nil
}
