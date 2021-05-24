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
	cfg.Set("listen_addr", "localhost:9001")
	cfg.Set("db_server", "redis:6379")
	cfg.Set("config_watch_interval", 5)
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

	t.Run("Update greeting", func(t *testing.T) {
		r, err := http.Post("http://localhost:9001/hello", "application/text", bytes.NewBuffer([]byte("Howdie")))
		if err != nil {
			t.Errorf("Unexpected error when updating greeting - %s", err)
		}
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when updating greeting - %d", r.StatusCode)
		}
	})

	t.Run("Check greeting", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/hello")
		if err != nil {
			t.Errorf("Unexpected error when requesting greeting service - %s", err)
		}
		if r.StatusCode != 200 {
			t.Errorf("Unexpected http status code when checking greeting service - %d", r.StatusCode)
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

	t.Run("Update greeting with DB Closed", func(t *testing.T) {
		r, err := http.Post("http://localhost:9001/hello", "application/text", bytes.NewBuffer([]byte("Howdie2")))
		if err != nil {
			t.Errorf("Unexpected error when updating greeting - %s", err)
		}
		if r.StatusCode != 500 {
			t.Errorf("Unexpected http status code when updating greeting - %d", r.StatusCode)
		}
	})

	t.Run("Check greeting with DB Closed", func(t *testing.T) {
		r, err := http.Get("http://localhost:9001/hello")
		if err != nil {
			t.Errorf("Unexpected error when requesting greeting service - %s", err)
		}
		if r.StatusCode != 500 {
			t.Errorf("Unexpected http status code when checking greeting service - %d", r.StatusCode)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected error reading http response - %s", err)
		}
		if string(body) == string([]byte("Howdie2")) {
			t.Errorf("Unexpected reply from http response - got %s", body)
		}
	})

}
