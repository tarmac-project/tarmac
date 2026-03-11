package main

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSetDefaults(t *testing.T) {
	cfg := viper.New()

	setDefaults(cfg)

	testCases := []struct {
		key  string
		want any
	}{
		{key: "enable_tls", want: true},
		{key: "listen_addr", want: "0.0.0.0:8443"},
		{key: "cert_file", want: "/certs/cert.crt"},
		{key: "key_file", want: "/certs/key.key"},
		{key: "config_watch_interval", want: 15},
		{key: "wasm_function", want: "/functions/tarmac.wasm"},
		{key: "wasm_function_config", want: "/functions/tarmac.json"},
		{key: "kvstore_type", want: "internal"},
		{key: "boltdb_filename", want: "/data/tarmac/tarmac.db"},
		{key: "boltdb_bucket", want: "tarmac"},
		{key: "boltdb_permissions", want: 0600},
		{key: "boltdb_timeout", want: 5},
		{key: "nats_url", want: "nats://localhost:4222"},
		{key: "nats_bucket", want: "tarmac"},
		{key: "grpc_socket_path", want: "/grpc.sock"},
		{key: "run_mode", want: "daemon"},
		{key: "text_log_format", want: false},
		{key: "http_client_max_response_body_size", want: 10 * 1024 * 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			if got := cfg.Get(tc.key); got != tc.want {
				t.Fatalf("unexpected default for %s: got %v want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestHandleConfigReadResult(t *testing.T) {
	var logBuffer bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{}))

	t.Run("nil error", func(t *testing.T) {
		if err := handleConfigReadResult(log, nil); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("config file not found", func(t *testing.T) {
		logBuffer.Reset()

		err := handleConfigReadResult(log, viper.ConfigFileNotFoundError{})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !strings.Contains(logBuffer.String(), "No Config file found") {
			t.Fatalf("expected warning log, got %q", logBuffer.String())
		}
	})

	t.Run("other error", func(t *testing.T) {
		wantErr := errors.New("boom")

		err := handleConfigReadResult(log, wantErr)
		if err == nil {
			t.Fatal("expected an error")
		}

		if !errors.Is(err, wantErr) {
			t.Fatalf("expected wrapped error, got %v", err)
		}
	})
}
