package app

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
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
	cfg.Set("wasm_function", "/testdata/base/default/tarmac.wasm")

	srv := New(cfg)
	go func() {
		err := srv.Run()
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

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

func TestMaintenanceMode(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", false)
	cfg.Set("debug", true)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("wasm_function", "/testdata/base/default/tarmac.wasm")
	cfg.Set("enable_maintenance_mode", true)

	srv := New(cfg)
	go func() {
		err := srv.Run()
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

	// Wait for Server to start
	time.Sleep(2 * time.Second)

	t.Run("Check Readiness", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/ready")
		if err != nil {
			t.Fatalf("Unexpected error when making HTTP request - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 503 {
			t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
		}
	})
}

type FullServiceTestCase struct {
	name string
	cfg  *viper.Viper
}

func TestFullService(t *testing.T) {
	var tt []FullServiceTestCase

	tc := FullServiceTestCase{name: "Redis", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("enable_tls", false)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "redis")
	tc.cfg.Set("redis_server", "redis:6379")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("wasm_function_config", "/testdata/tarmac.json")
	tt = append(tt, tc)

	tc = FullServiceTestCase{name: "In-Memory", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "in-memory")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("wasm_function_config", "/testdata/tarmac.json")
	tt = append(tt, tc)

	tc = FullServiceTestCase{name: "Cassandra", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "cassandra")
	tc.cfg.Set("cassandra_hosts", []string{"cassandra-primary", "cassandra"})
	tc.cfg.Set("cassandra_keyspace", "tarmac")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("wasm_function_config", "/testdata/tarmac.json")
	tt = append(tt, tc)

	fh, err := os.CreateTemp("", "*.db")
	if err != nil {
		t.Fatalf("Unexpected error creating temp file - %s", err)
	}
	defer os.Remove(fh.Name())
	fh.Close()

	tc = FullServiceTestCase{name: "BoltDB", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "internal")
	tc.cfg.Set("boltdb_filename", fh.Name())
	tc.cfg.Set("boltdb_bucket", "tarmac")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("wasm_function_config", "/testdata/tarmac.json")
	tt = append(tt, tc)

	tc = FullServiceTestCase{name: "MySQL", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("enable_sql", true)
	tc.cfg.Set("sql_type", "mysql")
	tc.cfg.Set("sql_dsn", "root:example@tcp(mysql:3306)/example")
	tc.cfg.Set("wasm_function_config", "/testdata/tarmac.json")
	tt = append(tt, tc)

	tc = FullServiceTestCase{name: "Postgres", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("enable_sql", true)
	tc.cfg.Set("sql_type", "postgres")
	tc.cfg.Set("sql_dsn", "postgres://example:example@postgres:5432/example?sslmode=disable")
	tc.cfg.Set("wasm_function_config", "/testdata/tarmac.json")
	tt = append(tt, tc)

	tc = FullServiceTestCase{name: "In-Memory SDKv1", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "in-memory")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("wasm_function_config", "/testdata/sdkv1/tarmac.json")
	tt = append(tt, tc)

	tc = FullServiceTestCase{name: "MySQL SDKv1", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("enable_sql", true)
	tc.cfg.Set("sql_type", "mysql")
	tc.cfg.Set("sql_dsn", "root:example@tcp(mysql:3306)/example")
	tc.cfg.Set("wasm_function_config", "/testdata/sdkv1/tarmac.json")
	tt = append(tt, tc)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			srv := New(tc.cfg)
			go func() {
				err := srv.Run()
				if err != nil && err != ErrShutdown {
					t.Errorf("Run unexpectedly stopped - %s", err)
				}
			}()
			// Clean up
			defer srv.Stop()

			// Wait for Server to start
			time.Sleep(30 * time.Second)

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

			if tc.cfg.GetBool("enable_kvstore") {
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
			}

			if tc.cfg.GetBool("enable_sql") {
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
			}

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

		})
	}
}

type InitFuncTestCase struct {
	name   string
	cfg    *viper.Viper
	config []byte
	err    bool
}

func TestInitFuncs(t *testing.T) {
	var tt []InitFuncTestCase

	tc := InitFuncTestCase{name: "Default", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "in-memory")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("run_mode", "job")
	tc.config = []byte(`{"services":{"test-service":{"name":"test-service","functions":{"default":{"filepath":"/testdata/base/default/tarmac.wasm","pool_size":1}},"routes":[{"type":"init","function":"default"}]}}}`)
	tt = append(tt, tc)

	tc = InitFuncTestCase{name: "Fails", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "in-memory")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("run_mode", "job")
	tc.config = []byte(`{"services":{"test-service":{"name":"test-service","functions":{"fail":{"filepath":"/testdata/base/fail/tarmac.wasm","pool_size":1}},"routes":[{"type":"init","function":"fail"}]}}}`)
	tc.err = true
	tt = append(tt, tc)

	tc = InitFuncTestCase{name: "Success After 5 Retries", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "in-memory")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("run_mode", "job")
	tc.config = []byte(`{"services":{"test-service":{"name":"test-service","functions":{"successafter5":{"filepath":"/testdata/base/successafter5/tarmac.wasm","pool_size":1}},"routes":[{"type":"init","retries":10,"function":"successafter5"}]}}}`)
	tt = append(tt, tc)

	tc = InitFuncTestCase{name: "Fail After 10 Retries", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "in-memory")
	tc.cfg.Set("enable_kvstore", true)
	tc.cfg.Set("run_mode", "job")
	tc.config = []byte(`{"services":{"test-service":{"name":"test-service","functions":{"fail":{"filepath":"/testdata/base/fail/tarmac.wasm","pool_size":1}},"routes":[{"type":"init","retries":10,"function":"fail"}]}}}`)
	tc.err = true
	tt = append(tt, tc)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Write the config to a temp file
			fh, err := os.CreateTemp("", "*.json")
			if err != nil {
				t.Fatalf("Unexpected error creating temp file - %s", err)
			}
			defer os.Remove(fh.Name())
			if _, err := fh.Write(tc.config); err != nil {
				t.Fatalf("Unexpected error writing to temp file - %s", err)
			}
			fh.Close()
			tc.cfg.Set("wasm_function_config", fh.Name())

			// Create the server
			srv := New(tc.cfg)
			defer srv.Stop()

			// Time out after 2 seconds
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			go func() {
				<-ctx.Done()
				defer srv.Stop()
				if ctx.Err() == context.DeadlineExceeded && tc.err {
					t.Errorf("Timeout waiting for server to start")
				}
			}()

			// Start the server
			err = srv.Run()
			if err != nil && err != ErrShutdown {
				if tc.err {
					return
				}
				t.Errorf("Run unexpectedly stopped - %s", err)
			}
			if err == ErrShutdown && ctx.Err() == context.DeadlineExceeded && !tc.err {
				t.Errorf("Server did not start and shutdown as expected")
			}

			if ctx.Err() == context.DeadlineExceeded && tc.err {
				t.Fatalf("Server did not fail as expected")
			}

		})
	}
}

func TestWASMRunner(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", false)
	cfg.Set("debug", true)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("wasm_function", "/testdata/base/default/tarmac.wasm")
	srv := New(cfg)
	go func() {
		err := srv.Run()
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

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
			_, err := srv.runWASM(c.module, c.handler, c.request)
			if err != nil && !c.err {
				t.Errorf("Unexpected error executing module - %s", err)
			}
			if err == nil && c.err {
				t.Errorf("Unexpected success executing module")
			}
		})
	}

}
