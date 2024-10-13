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
	v.Set("wasm_function_config", "/testdata/tarmac-fail.json")
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
			if err == nil || err == ErrShutdown {
				t.Errorf("Expected error when starting server, got nil")
			}
		})
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
	time.Sleep(10 * time.Second)

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

func TestPProfServerEnabled(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", true)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("config_watch_interval", 5)
	cfg.Set("use_consul", false)
	cfg.Set("debug", true)
	cfg.Set("trace", true)
	cfg.Set("enable_pprof", true)
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
	time.Sleep(10 * time.Second)

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
	time.Sleep(10 * time.Second)

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
			if r.StatusCode != 403 {
				t.Errorf("Unexpected http status code when validating pprof - %d", r.StatusCode)
			}
		})
	}
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
	time.Sleep(15 * time.Second)

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
	time.Sleep(15 * time.Second)

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
	time.Sleep(15 * time.Second)

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/health")
		if err == nil {
			defer r.Body.Close()
			t.Errorf("Unexpected success when requesting health status")
			t.FailNow()
		}
	})
}
