package app

import (
	"bytes"
	"github.com/spf13/viper"
	"io/ioutil"
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

func TestHandlers(t *testing.T) {
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
	cfg.Set("wasm_function", "/testdata/tarmac.wasm")
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
		body, err := ioutil.ReadAll(r.Body)
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
		if r.StatusCode != 500 {
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
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected error reading http response - %s", err)
		}
		if string(body) != string([]byte("Howdie")) {
			t.Errorf("Unexpected reply from http response - got %s", body)
		}
	})

	t.Run("Invalid Head Request", func(t *testing.T) {
		r, err := http.Head("http://localhost:9001/")
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode < 500 {
			t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
		}
	})
}

func TestWASMRunner(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", false)
	cfg.Set("debug", true)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("kvstore_type", "redis")
	cfg.Set("redis_server", "redis:6379")
	cfg.Set("enable_kvstore", true)
	cfg.Set("config_watch_interval", 5)
	cfg.Set("enable_sql", true)
	cfg.Set("sql_type", "mysql")
	cfg.Set("sql_dsn", "root:example@tcp(mysql:3306)/example")
	cfg.Set("wasm_function", "/testdata/tarmac.wasm")
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
		handler: "GET",
		request: []byte(""),
	})

	tc = append(tc, RunnerCase{
		name:    "Happy Path - Bad Payload",
		err:     false,
		pass:    false,
		module:  "default",
		handler: "POST",
		request: []byte("ohmy"),
	})

	tc = append(tc, RunnerCase{
		name:    "Happy Path",
		err:     false,
		pass:    true,
		module:  "default",
		handler: "POST",
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
			rsp, err := runWASM(c.module, c.handler, c.request)
			if err != nil && !c.err {
				t.Errorf("Unexpected error executing module - %s", err)
			}
			if err == nil && c.err {
				t.Errorf("Unexpected success executing module")
			}

			t.Logf("%s", rsp)
		})
	}

}
