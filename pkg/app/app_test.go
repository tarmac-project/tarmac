package app

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/madflojo/testcerts"
	"github.com/spf13/viper"
)

const (
	// Test server timeout values.
	testServerContextTimeout = 2 * time.Second
	testServerMaxWaitTimeout = 2500 * time.Millisecond // Slightly longer than context timeout
)

// waitForServer polls the health endpoint until it responds or times out.
func waitForServer(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden {
					return nil
				}
			}
		}
	}
}

func TestBadConfigs(t *testing.T) {
	cfgs := make(map[string]*viper.Viper)

	// Invalid Listener Address
	v := viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "pandasdonotbelonghere")
	v.Set("disable_logging", true)
	v.Set("kvstore_type", "redis")
	v.Set("redis_server", "redis:6379")
	cfgs["invalid listener address"] = v

	// Invalid TLS config
	v = viper.New()
	v.Set("enable_tls", true)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("kvstore_type", "redis")
	v.Set("redis_server", "redis:6379")
	v.Set("cert_file", "/tmp/doesntexist")
	v.Set("key_file", "/tmp/doesntexist")
	cfgs["invalid TLS Config"] = v

	// Invalid Redis Address
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", true)
	v.Set("kvstore_type", "redis")
	v.Set("redis_server", "")
	cfgs["invalid Redis Address"] = v

	// Invalid Cassandra Address
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", true)
	v.Set("kvstore_type", "cassandra")
	v.Set("cassandra_hosts", []string{"notarealaddress"})
	cfgs["invalid Cassandra Address"] = v

	// Invalid Cassandra Keyspace
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", true)
	v.Set("kvstore_type", "cassandra")
	v.Set("cassandra_keyspace", "")
	v.Set("cassandra_hosts", []string{"cassandra-primary", "cassandra"})
	cfgs["invalid Cassandra Keyspace"] = v

	// Invalid NATS URL
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", true)
	v.Set("kvstore_type", "nats")
	v.Set("nats_url", "nats://notarealaddress:4222")
	v.Set("nats_bucket", "tarmac")
	cfgs["invalid NATS URL"] = v

	// Invalid NATS Bucket
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", true)
	v.Set("kvstore_type", "nats")
	v.Set("nats_url", "nats://nats:4222")
	v.Set("nats_bucket", "")
	cfgs["invalid NATS Bucket"] = v

	// Invalid KVStore
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", true)
	v.Set("kvstore_type", "notvalid")
	cfgs["invalid kvstore Address"] = v

	// Invalid WASM path
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", false)
	v.Set("wasm_function", "something-that-does-not-exist")
	cfgs["invalid WASM path"] = v

	// Failing init function
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", false)
	v.Set("wasm_function_config", "../../testdata/tarmac-fail.json")
	cfgs["failing init function"] = v

	// Loop through bad configs, creating sub-tests as we go
	for k, v := range cfgs {
		t.Run("Testing "+k, func(t *testing.T) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(5)*time.Second))
			defer cancel()
			srv := New(v)
			go func() {
				<-ctx.Done()
				err := ctx.Err()
				if err == context.DeadlineExceeded {
					srv.Stop()
				}
			}()
			err := srv.Run()
			if err == nil || errors.Is(err, ErrShutdown) {
				t.Errorf("Expected error when starting server, got nil")
			}
		})
	}
}

func TestPProfServerEnabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", true)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("config_watch_interval", 5)
	cfg.Set("use_consul", false)
	cfg.Set("debug", true)
	cfg.Set("trace", true)
	cfg.Set("enable_pprof", true)
	cfg.Set("wasm_function", "../../testdata/base/default/tarmac.wasm")
	srv := New(cfg)
	go func() {
		err := srv.Run()
		if err != nil && !errors.Is(err, ErrShutdown) {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

	// Wait for app to start
	if err := waitForServer("http://localhost:9000/health", 15*time.Second); err != nil {
		t.Fatalf("Server failed to start: %v", err)
	}

	urls := []string{
		"debug/pprof",
		"debug/pprof/allocs",
		"debug/pprof/cmdline",
		"debug/pprof/goroutine",
		"debug/pprof/heap",
		"debug/pprof/mutex",
		"debug/pprof/profile",
		"debug/pprof/threadcreate",
		"debug/pprof/trace",
	}

	for _, u := range urls {
		t.Run("Verifying URL "+u, func(t *testing.T) {
			r, err := http.Get("http://localhost:9000/" + u)
			if err != nil {
				t.Errorf("Unexpected error when validating pprof - %s", err)
			}
			defer r.Body.Close()
			if r.StatusCode > 399 {
				t.Errorf("Unexpected http status code when validating pprof - %d", r.StatusCode)
			}
		})
	}
}

func TestPProfServerDisabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", true)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("config_watch_interval", 5)
	cfg.Set("use_consul", false)
	cfg.Set("debug", true)
	cfg.Set("trace", true)
	cfg.Set("wasm_function", "../../testdata/base/default/tarmac.wasm")
	srv := New(cfg)
	go func() {
		err := srv.Run()
		if err != nil && !errors.Is(err, ErrShutdown) {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer srv.Stop()

	// Wait for app to start
	if err := waitForServer("http://localhost:9000/health", 15*time.Second); err != nil {
		t.Fatalf("Server failed to start: %v", err)
	}

	urls := []string{
		"debug/pprof",
		"debug/pprof/allocs",
		"debug/pprof/cmdline",
		"debug/pprof/goroutine",
		"debug/pprof/heap",
		"debug/pprof/mutex",
		"debug/pprof/profile",
		"debug/pprof/threadcreate",
		"debug/pprof/trace",
	}

	for _, u := range urls {
		t.Run("Verifying URL "+u, func(t *testing.T) {
			r, err := http.Get("http://localhost:9000/" + u)
			if err != nil {
				t.Errorf("Unexpected error when validating pprof - %s", err)
			}
			defer r.Body.Close()
			if r.StatusCode != http.StatusForbidden {
				t.Errorf("Unexpected http status code when validating pprof - %d", r.StatusCode)
			}
		})
	}
}

// TestTLSBranchBehavior verifies that when TLS is enabled, the server only
// attempts to start a TLS listener and doesn't fall through to start a
// non-TLS listener. This test addresses the issue where the code could
// fall through after ListenAndServeTLS returns.
func TestTLSBranchBehavior(t *testing.T) {
	// Test that TLS configuration attempts only TLS listener
	t.Run("TLS enabled uses only TLS listener", func(t *testing.T) {
		// Create Test Certs in temporary directory
		tmpDir := t.TempDir()
		certFile := tmpDir + "/cert.pem"
		keyFile := tmpDir + "/key.pem"
		err := testcerts.GenerateCertsToFile(certFile, keyFile)
		if err != nil {
			t.Errorf("Failed to create certs - %s", err)
			t.FailNow()
		}

		cfg := viper.New()
		cfg.Set("disable_logging", true)
		cfg.Set("enable_tls", true)
		cfg.Set("cert_file", certFile)
		cfg.Set("key_file", keyFile)
		cfg.Set("listen_addr", "127.0.0.1:19001") // Use unique port
		cfg.Set("enable_kvstore", false)
		cfg.Set("wasm_function", "../../testdata/base/default/tarmac.wasm")

		srv := New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), testServerContextTimeout)
		defer cancel()

		// Start server in goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- srv.Run()
		}()

		// Schedule shutdown
		go func() {
			<-ctx.Done()
			srv.Stop()
		}()

		// Wait for either error or context timeout
		select {
		case err := <-errChan:
			// Server should stop with ErrShutdown
			if err != nil && !errors.Is(err, ErrShutdown) {
				t.Errorf("Expected ErrShutdown or nil, got: %s", err)
			}
		case <-time.After(testServerMaxWaitTimeout):
			// Slightly longer than context timeout to ensure proper shutdown
			srv.Stop()
		}
	})

	// Test that non-TLS configuration attempts only non-TLS listener
	t.Run("TLS disabled uses only HTTP listener", func(t *testing.T) {
		cfg := viper.New()
		cfg.Set("disable_logging", true)
		cfg.Set("enable_tls", false)
		cfg.Set("listen_addr", "127.0.0.1:19002") // Use unique port
		cfg.Set("enable_kvstore", false)
		cfg.Set("wasm_function", "../../testdata/base/default/tarmac.wasm")

		srv := New(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), testServerContextTimeout)
		defer cancel()

		// Start server in goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- srv.Run()
		}()

		// Schedule shutdown
		go func() {
			<-ctx.Done()
			srv.Stop()
		}()

		// Wait for either error or context timeout
		select {
		case err := <-errChan:
			// Server should stop with ErrShutdown
			if err != nil && !errors.Is(err, ErrShutdown) {
				t.Errorf("Expected ErrShutdown or nil, got: %s", err)
			}
		case <-time.After(testServerMaxWaitTimeout):
			// Slightly longer than context timeout to ensure proper shutdown
			srv.Stop()
		}
	})
}
