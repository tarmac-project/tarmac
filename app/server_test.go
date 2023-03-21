package app

import (
	"bytes"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"testing"
	"time"
)

type RunnerCase struct {
	name    string
	err     bool
	pass    bool
	module  string
	handler string
	request []byte
}

func TestBasicFunction(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", false)
	cfg.Set("debug", true)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("wasm_function", "/testdata/default/tarmac.wasm")
	go func() {
		err := Run(cfg)
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer Stop()

	// Wait for Server to start
	time.Sleep(2 * time.Second)

	t.Run("Simple Payload", func(t *testing.T) {
		r, err := http.Post("http://localhost:9001/", "application/text", bytes.NewBuffer([]byte("Howdie")))
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making HTTP request %d", r.StatusCode)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected error reading http response - %s", err)
		}
		if string(body) != string([]byte("Howdie")) {
			t.Errorf("Unexpected reply from http response - got %s", body)
		}
	})

	t.Run("Do a Get", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/")
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
		}
	})

	t.Run("No Payload", func(t *testing.T) {
		r, err := http.Post("http://localhost:9001/", "application/text", bytes.NewBuffer([]byte("")))
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making HTTP request %d", r.StatusCode)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected error reading http response - %s", err)
		}
		if string(body) != string([]byte("Howdie")) {
			t.Errorf("Unexpected reply from http response - got %s", body)
		}
	})

}

func TestFullService(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", false)
	cfg.Set("debug", true)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("kvstore_type", "redis")
	cfg.Set("redis_server", "redis:6379")
	cfg.Set("enable_kvstore", true)
	cfg.Set("enable_sql", true)
	cfg.Set("sql_type", "mysql")
	cfg.Set("sql_dsn", "root:example@tcp(mysql:3306)/example")
	cfg.Set("config_watch_interval", 5)
	cfg.Set("wasm_function_config", "/testdata/tarmac.json")
	cfg.Set("wasm_function", "/testdata/doesnotexist/tarmac.wasm")
	go func() {
		err := Run(cfg)
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer Stop()

	// Wait for Server to start
	time.Sleep(2 * time.Second)

	// Call /logger with POST
	t.Run("Do a Post on /logger", func(t *testing.T) {
		r, err := http.Post("http://localhost:9001/logger", "application/text", bytes.NewBuffer([]byte("Test Payload")))
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making HTTP request %d", r.StatusCode)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected error reading http response - %s", err)
		}
		if string(body) != string([]byte("Test Payload")) {
			t.Errorf("Unexpected reply from http response - got %s", body)
		}
	})

	// Call /kv and /sql with GET
	t.Run("Do a Get on /kv", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/kv")
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
		}
	})

	t.Run("Do a Get on /sql", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/sql")
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
		}
	})

	t.Run("Do a Get on /func", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/func")
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
		}
	})
}

func TestWASMRunner(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", false)
	cfg.Set("debug", true)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("wasm_function", "/testdata/default/tarmac.wasm")
	go func() {
		err := Run(cfg)
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer Stop()

	// Wait for Server to start
	time.Sleep(2 * time.Second)

	var tc []RunnerCase

	tc = append(tc, RunnerCase{
		name:    "Module Doesn't Exist",
		err:     true,
		pass:    false,
		module:  "notfound",
		handler: "handler",
		request: []byte(""),
	})

	tc = append(tc, RunnerCase{
		name:    "Happy Path",
		err:     false,
		pass:    true,
		module:  "default",
		handler: "handler",
		request: []byte("howdie"),
	})

	tc = append(tc, RunnerCase{
		name:    "Bad Handler Route",
		err:     true,
		pass:    false,
		module:  "default",
		handler: "noroute",
		request: []byte(""),
	})

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			_, err := runWASM(c.module, c.handler, c.request)
			if err != nil && !c.err {
				t.Errorf("Unexpected error executing module - %s", err)
			}
			if err == nil && c.err {
				t.Errorf("Unexpected success executing module")
			}
		})
	}

}
