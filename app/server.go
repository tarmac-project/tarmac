package app

import (
	//	"encoding/base64"
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

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
	err := kv.HealthCheck()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// middleware is used to intercept incoming HTTP calls and apply general functions upon
// them. e.g. Metrics, Logging...
func (s *server) middleware(n httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Fetch Module and run with payload
	m, err := engine.Module("default")
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error loading wasi environment %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Execute the WASM HTTP Handler
	rsp, err := m.Run("request:handler", payload)
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote-addr":    r.RemoteAddr,
			"http-protocol":  r.Proto,
			"headers":        r.Header,
			"content-length": r.ContentLength,
		}).Debugf("Error executing function %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return status code and print stdout
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", rsp)
}

func (s *server) CallbackRouter(ctx context.Context, binding, namespace, op string, payload []byte) ([]byte, error) {
	log.WithFields(logrus.Fields{
		"binding":   binding,
		"namespace": namespace,
		"function":  op,
	}).Infof("CallbackRouter called with payload %s", payload)
	return []byte(""), nil
}
