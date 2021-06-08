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
	cfg.Set("disable_logging", true)
	cfg.Set("debug", false)
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("db_server", "redis:6379")
	cfg.Set("enable_db", true)
	cfg.Set("config_watch_interval", 5)
	cfg.Set("wasm_module", "/example/go/http_env/module/tarmac_module.wasm")
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

	t.Run("Do a Get", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/")
		if err != nil {
			t.Errorf("Unexpected error when making HTTP request - %s", err)
		}
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when making request %d", r.StatusCode)
		}
	})

	t.Run("No Payload", func(t *testing.T) {
		r, err := http.Post("http://localhost:9001/", "application/text", bytes.NewBuffer([]byte("")))
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
		if string(body) != string([]byte("")) {
			t.Errorf("Unexpected reply from http response - got %s", body)
		}
	})
}
