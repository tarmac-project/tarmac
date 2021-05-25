package app

import (
	"bytes"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestHandlers(t *testing.T) {
	cfg := viper.New()
	cfg.Set("disable_logging", false)
  cfg.Set("debug", true)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("db_server", "redis:6379")
	cfg.Set("enable_db", true)
	cfg.Set("config_watch_interval", 5)
	cfg.Set("wasm-module", "example/go/tarmac-module.wasm")
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

	t.Run("Simple Payload", func(t *testing.T) {
		r, err := http.Post("http://localhost:9001/", "application/text", bytes.NewBuffer([]byte("Howdie")))
		if err != nil {
			t.Errorf("Unexpected error when making HTTP request - %s", err)
		}
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

	// Close DB for error checks
	db.Close()
}
