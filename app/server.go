package app

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/madflojo/tarmac/pkg/config"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
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

	// funcCfg is used to store and access multi-function service configurations
	funcCfg *config.Config
}

// Health is used to handle HTTP Health requests to this service. Use this for liveness
// probes or any other checks which only validate if the services is running.
func (s *server) Health(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}

// Ready is used to handle HTTP Ready requests to this service. Use this for readiness
// probes or any checks that validate the service is ready to accept traffic.
func (s *server) Ready(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
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

			stats.Srv.WithLabelValues(r.URL.Path).Observe(time.Since(now).Seconds())
			return
		}

		// Call registered handler
		n(w, r, ps)
		stats.Srv.WithLabelValues(r.URL.Path).Observe(time.Since(now).Seconds())
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
func (s *server) WASMHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Find Function
	function, err := s.funcCfg.RouteLookup(fmt.Sprintf("http:%s:%s", r.Method, r.URL.Path))
	if err == config.ErrRouteNotFound {
		function = "default"
	}

	// Read the HTTP Payload
	var payload []byte
	if r.Method == "POST" || r.Method == "PUT" {
		payload, err = io.ReadAll(r.Body)
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

	// Execute WASM Module
	rsp, err := runWASM(function, "handler", payload)
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

	// Return status code and print stdout
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", rsp)
}

// runWASM will load and execute the specified WASM module.
func runWASM(module, handler string, rq []byte) ([]byte, error) {
	now := time.Now()

	// Fetch Module and run with payload
	m, err := engine.Module(module)
	if err != nil {
		stats.Wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
		return []byte(""), fmt.Errorf("unable to load wasi environment - %s", err)
	}

	// Execute the WASM Handler
	rsp, err := m.Run(handler, rq)
	if err != nil {
		stats.Wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
		return rsp, fmt.Errorf("failed to execute wasm module - %s", err)
	}

	// Return results
	stats.Wasm.WithLabelValues(fmt.Sprintf("%s:%s", module, handler)).Observe(time.Since(now).Seconds())
	return rsp, nil
}
