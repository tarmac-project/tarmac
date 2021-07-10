package app

import (
	"context"
	"crypto/tls"
	"github.com/madflojo/testcerts"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestBadConfigs(t *testing.T) {
	cfgs := make(map[string]*viper.Viper)

	// Invalid Listener Address
	v := viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "pandasdonotbelonghere")
	v.Set("disable_logging", true)
	v.Set("kv_server", "redis:6379")
	cfgs["invalid listener address"] = v

	// Invalid TLS config
	v = viper.New()
	v.Set("enable_tls", true)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("kv_server", "redis:6379")
	v.Set("cert_file", "/tmp/doesntexist")
	v.Set("key_file", "/tmp/doesntexist")
	cfgs["invalid TLS Config"] = v

	// Invalid KV Address
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", true)
	v.Set("kv_server", "")
	cfgs["invalid KV Address"] = v

	// Invalid WASM path
	v = viper.New()
	v.Set("enable_tls", false)
	v.Set("listen_addr", "0.0.0.0:8443")
	v.Set("disable_logging", true)
	v.Set("enable_kvstore", false)
	v.Set("wasm_function", "something-that-does-not-exist")
	cfgs["invalid WASM path"] = v

	// Loop through bad configs, creating sub-tests as we go
	for k, v := range cfgs {
		t.Run("Testing "+k, func(t *testing.T) {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(5)*time.Second))
			defer cancel()
			go func() {
				<-ctx.Done()
				err := ctx.Err()
				if err == context.DeadlineExceeded {
					Stop()
				}
			}()
			err := Run(v)
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
	cfg.Set("kv_server", "redis:6379")
	cfg.Set("enable_kvstore", true)
	cfg.Set("config_watch_interval", 5)
	cfg.Set("use_consul", false)
	cfg.Set("debug", true)
	cfg.Set("trace", true)
	go func() {
		err := Run(cfg)
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer Stop()

	// Wait for app to start
	time.Sleep(10 * time.Second)

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("http://localhost:9000/health")
		if err != nil {
			t.Errorf("Unexpected error when requesting health status - %s", err)
		}
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking health - %d", r.StatusCode)
		}
	})

	/*
		t.Run("Check Scheduler is set", func(t *testing.T) {
			if len(scheduler.Tasks()) == 0 {
				t.Errorf("Expected scheduler to have at least one task")
			}
		})
	*/
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
	cfg.Set("enable_tls", true)
	cfg.Set("cert_file", "/tmp/cert")
	cfg.Set("key_file", "/tmp/key")
	cfg.Set("kv_server", "redis:6379")
	cfg.Set("enable_kvstore", true)
	cfg.Set("listen_addr", "localhost:9000")
	cfg.Set("config_watch_interval", 1)
	err = cfg.AddRemoteProvider("consul", "consul:8500", "tarmac/config")
	if err != nil {
		t.Fatalf("Failed to create Consul config provider - %s", err)
	}
	cfg.SetConfigType("json")
	_ = cfg.ReadRemoteConfig()

	// Start Server in goroutine
	go func() {
		err := Run(cfg)
		if err != nil && err != ErrShutdown {
			t.Errorf("Run unexpectedly stopped - %s", err)
		}
	}()
	// Clean up
	defer Stop()

	// Wait for app to start
	time.Sleep(10 * time.Second)

	t.Run("Check Health HTTP Handler", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/health")
		if err != nil {
			t.Errorf("Unexpected error when requesting health status - %s", err)
			t.FailNow()
		}
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
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking readiness - %d", r.StatusCode)
		}
	})

	// Kill the DB sessions for unhappy path testing
	kv.Close()

	t.Run("Check Ready HTTP Handler with DB Stopped", func(t *testing.T) {
		r, err := http.Get("https://localhost:9000/ready")
		if err != nil {
			t.Errorf("Unexpected error when requesting ready status - %s", err)
			t.FailNow()
		}
		if r.StatusCode != 503 {
			t.Errorf("Unexpected http status code when checking readiness - %d", r.StatusCode)
		}
	})

}
