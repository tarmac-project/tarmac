//go:build integration

package app

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/madflojo/testcerts"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"github.com/tarmac-project/tarmac/pkg/tlsconfig"
)

// waitForServer polls the health endpoint until it responds or times out
func waitForServer(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					return nil
				}
			}
		}
	}
}

func TestRunningServer(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", true)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("kvstore_type", "redis")
	cfg.Set("redis_server", "redis:6379")
	cfg.Set("enable_kvstore", true)
	cfg.Set("config_watch_interval", 5)
	cfg.Set("use_consul", false)
	cfg.Set("debug", true)
	cfg.Set("trace", true)
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

	// Wait for app to start
	if err := waitForServer("http://localhost:9000/health", 15*time.Second); err != nil {
		t.Fatalf("Server failed to start: %v", err)
	}

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("http://localhost:9000/health")
		if err != nil {
			t.Errorf("Unexpected error when requesting health status - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking health - %d", r.StatusCode)
		}
	})

	t.Run("Check Metrics HTTP Handler", func(t *testing.T) {
		r, err := http.Get("http://localhost:9000/metrics")
		if err != nil {
			t.Errorf("Unexpected error when requesting metrics status - %s", err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking metrics - %d", r.StatusCode)
		}
	})
}

func TestRunningTLSServer(t *testing.T) {
	// Create Test Certs
	err := testcerts.GenerateCertsToFile("/tmp/cert", "/tmp/key")
	if err != nil {
		t.Errorf("Failed to create certs - %s", err)
		t.FailNow()
	}
	defer os.Remove("/tmp/cert")
	defer os.Remove("/tmp/key")

	// Disable Host Checking globally
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Setup Config
	cfg := viper.New()
	cfg.Set("disable_logging", true)
	cfg.Set("trace", true)
	cfg.Set("debug", true)
	cfg.Set("enable_tls", true)
	cfg.Set("cert_file", "/tmp/cert")
	cfg.Set("key_file", "/tmp/key")
	cfg.Set("kvstore_type", "redis")
	cfg.Set("redis_server", "redis:6379")
	cfg.Set("enable_kvstore", true)
	cfg.Set("use_consul", false)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("config_watch_interval", 1)
	cfg.Set("enable_sql", true)
	cfg.Set("sql_type", "mysql")
	cfg.Set("sql_dsn", "root:example@tcp(mysql:3306)/example")
	cfg.Set("wasm_function", "/testdata/base/default/tarmac.wasm")
	err = cfg.AddRemoteProvider("consul", "consul:8500", "tarmac/config")
	if err != nil {
		t.Fatalf("Failed to create Consul config provider - %s", err)
	}
	cfg.SetConfigType("json")
	_ = cfg.ReadRemoteConfig()

	srv := New(cfg)
	// Start Server in goroutine
	go func() {
		err := srv.Run()
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

	// Wait for app to start
	if err := waitForServer("https://localhost:9000/health", 20*time.Second); err != nil {
		t.Fatalf("Server failed to start: %v", err)
	}

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/health")
		if err != nil {
			t.Errorf("Unexpected error when requesting health status - %s", err)
			t.FailNow()
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking health - %d", r.StatusCode)
		}
	})

	t.Run("Check Ready HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/ready")
		if err != nil {
			t.Errorf("Unexpected error when requesting ready status - %s", err)
			t.FailNow()
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking readiness - %d", r.StatusCode)
		}
	})

	// Kill the DB sessions for unhappy path testing
	srv.kv.Close()

	t.Run("Check Ready HTTP Handler with DB Stopped", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/ready")
		if err != nil {
			t.Errorf("Unexpected error when requesting ready status - %s", err)
			t.FailNow()
		}
		defer r.Body.Close()
		if r.StatusCode != 503 {
			t.Errorf("Unexpected http status code when checking readiness - %d", r.StatusCode)
		}
	})

	t.Run("Check if Remote config was read", func(t *testing.T) {
		if !cfg.GetBool("from_consul") {
			t.Errorf("Did not fetch config from consul")
		}
	})

}

func TestRunningMTLSServer(t *testing.T) {
	// Create Test Certs
	err := testcerts.GenerateCertsToFile("/tmp/cert", "/tmp/key")
	if err != nil {
		t.Errorf("Failed to create certs - %s", err)
		t.FailNow()
	}
	defer os.Remove("/tmp/cert")
	defer os.Remove("/tmp/key")

	// Setup TLS Config
	tlsCfg := tlsconfig.New()
	err = tlsCfg.CertsFromFile("/tmp/cert", "/tmp/key")
	if err != nil {
		t.Fatalf("Failed to load certs - %s", err)
	}

	tlsCfg.IgnoreHostValidation()

	// Disable Host Checking globally
	http.DefaultTransport.(*http.Transport).TLSClientConfig = tlsCfg.Generate()

	// Setup Config
	cfg := viper.New()
	cfg.Set("disable_logging", true)
	cfg.Set("trace", true)
	cfg.Set("debug", true)
	cfg.Set("enable_tls", true)
	cfg.Set("cert_file", "/tmp/cert")
	cfg.Set("ca_file", "/tmp/cert")
	cfg.Set("key_file", "/tmp/key")
	cfg.Set("kvstore_type", "redis")
	cfg.Set("redis_server", "redis:6379")
	cfg.Set("enable_kvstore", true)
	cfg.Set("use_consul", false)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("config_watch_interval", 1)
	cfg.Set("enable_sql", true)
	cfg.Set("sql_type", "mysql")
	cfg.Set("sql_dsn", "root:example@tcp(mysql:3306)/example")
	cfg.Set("wasm_function", "/testdata/base/default/tarmac.wasm")
	err = cfg.AddRemoteProvider("consul", "consul:8500", "tarmac/config")
	if err != nil {
		t.Fatalf("Failed to create Consul config provider - %s", err)
	}
	cfg.SetConfigType("json")
	_ = cfg.ReadRemoteConfig()

	srv := New(cfg)
	// Start Server in goroutine
	go func() {
		err := srv.Run()
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

	// Wait for app to start
	if err := waitForServer("https://localhost:9000/health", 20*time.Second); err != nil {
		t.Fatalf("Server failed to start: %v", err)
	}

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/health")
		if err != nil {
			t.Errorf("Unexpected error when requesting health status - %s", err)
			t.FailNow()
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking health - %d", r.StatusCode)
		}
	})

	t.Run("Check Ready HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/ready")
		if err != nil {
			t.Errorf("Unexpected error when requesting ready status - %s", err)
			t.FailNow()
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking readiness - %d", r.StatusCode)
		}
	})

	// Kill the DB sessions for unhappy path testing
	srv.kv.Close()

	t.Run("Check Ready HTTP Handler with DB Stopped", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/ready")
		if err != nil {
			t.Errorf("Unexpected error when requesting ready status - %s", err)
			t.FailNow()
		}
		defer r.Body.Close()
		if r.StatusCode != 503 {
			t.Errorf("Unexpected http status code when checking readiness - %d", r.StatusCode)
		}
	})

	t.Run("Check if Remote config was read", func(t *testing.T) {
		if !cfg.GetBool("from_consul") {
			t.Errorf("Did not fetch config from consul")
		}
	})

}

func TestRunningFailMTLSServer(t *testing.T) {
	// Create Test Certs
	err := testcerts.GenerateCertsToFile("/tmp/cert", "/tmp/key")
	if err != nil {
		t.Errorf("Failed to create certs - %s", err)
		t.FailNow()
	}
	defer os.Remove("/tmp/cert")
	defer os.Remove("/tmp/key")

	// Setup TLS Config
	tlsCfg := tlsconfig.New()
	tlsCfg.IgnoreHostValidation()

	// Disable Host Checking globally
	http.DefaultTransport.(*http.Transport).TLSClientConfig = tlsCfg.Generate()

	// Setup Config
	cfg := viper.New()
	cfg.Set("disable_logging", true)
	cfg.Set("trace", true)
	cfg.Set("debug", true)
	cfg.Set("enable_tls", true)
	cfg.Set("cert_file", "/tmp/cert")
	cfg.Set("ca_file", "/tmp/cert")
	cfg.Set("key_file", "/tmp/key")
	cfg.Set("enable_kvstore", true)
	cfg.Set("kvstore_type", "redis")
	cfg.Set("redis_server", "redis:6379")
	cfg.Set("use_consul", false)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("config_watch_interval", 1)
	cfg.Set("enable_sql", true)
	cfg.Set("sql_type", "mysql")
	cfg.Set("sql_dsn", "root:example@tcp(mysql:3306)/example")
	cfg.Set("wasm_function", "/testdata/base/default/tarmac.wasm")
	err = cfg.AddRemoteProvider("consul", "consul:8500", "tarmac/config")
	if err != nil {
		t.Fatalf("Failed to create Consul config provider - %s", err)
	}
	cfg.SetConfigType("json")
	_ = cfg.ReadRemoteConfig()

	srv := New(cfg)
	// Start Server in goroutine
	go func() {
		err := srv.Run()
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

	// Wait for app to start
	if err := waitForServer("https://localhost:9000/health", 20*time.Second); err != nil {
		t.Fatalf("Server failed to start: %v", err)
	}

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/health")
		if err == nil {
			defer r.Body.Close()
			t.Errorf("Unexpected success when requesting health status")
			t.FailNow()
		}
	})
}
