//go:build integration

package app

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

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

	tc = FullServiceTestCase{name: "NATS", cfg: viper.New()}
	tc.cfg.Set("disable_logging", false)
	tc.cfg.Set("debug", true)
	tc.cfg.Set("enable_tls", false)
	tc.cfg.Set("listen_addr", "localhost:9001")
	tc.cfg.Set("kvstore_type", "nats")
	tc.cfg.Set("nats_url", "nats://nats:4222")
	tc.cfg.Set("nats_bucket", "tarmac")
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

	fh, err := os.CreateTemp(t.TempDir(), "*.db")
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
				if err != nil && !errors.Is(err, ErrShutdown) {
					t.Errorf("Run unexpectedly stopped - %s", err)
				}
			}()
			defer srv.Stop()

			if err := waitForIntegrationServer("http://localhost:9001/health", 30*time.Second); err != nil {
				t.Fatalf("Server failed to start: %v", err)
			}

			t.Run("Do a Post on /logger", func(t *testing.T) {
				r, err := http.Post(
					"http://localhost:9001/logger",
					"application/text",
					bytes.NewBufferString("Test Payload"),
				)
				if err != nil {
					t.Fatalf("Unexpected error when making HTTP request - %s", err)
				}
				defer r.Body.Close()
				if r.StatusCode != http.StatusOK {
					t.Errorf("Unexpected http status code when making HTTP request %d", r.StatusCode)
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("Unexpected error reading http response - %s", err)
				}
				if string(body) != "Test Payload" {
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
					if r.StatusCode != http.StatusOK {
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
					if r.StatusCode != http.StatusOK {
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
				if r.StatusCode != http.StatusOK {
					t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
				}
			})
		})
	}
}
