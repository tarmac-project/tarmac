package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestHostCallCompatibility(t *testing.T) {
	const (
		baseURL      = "http://localhost:19003"
		wantResponse = "kvlogging-ok"
	)
	fixturePath, err := filepath.Abs("../../testdata/hostcalls/kvlogging/tarmac.wasm")
	if err != nil {
		t.Fatalf("resolve raw hostcall fixture path: %v", err)
	}
	configPath := writeHostCallConfig(t, fixturePath)

	cfg := viper.New()
	cfg.Set("disable_logging", false)
	cfg.Set("listen_addr", "localhost:19003")
	cfg.Set("kvstore_type", "in-memory")
	cfg.Set("enable_kvstore", true)
	cfg.Set("use_consul", false)
	cfg.Set("wasm_function_config", configPath)

	srv := New(cfg)
	runErr := make(chan error, 1)
	go func() {
		runErr <- srv.Run()
	}()
	t.Cleanup(func() {
		srv.Stop()
	})

	client := &http.Client{Timeout: time.Second}
	if err := waitForHostCallServer(client, baseURL+"/health", runErr); err != nil {
		t.Fatal(err)
	}

	response, err := client.Get(baseURL + "/kvlogging")
	if err != nil {
		t.Fatalf("invoke raw hostcall fixture: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read raw hostcall fixture response: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("raw hostcall fixture status = %d, body = %q", response.StatusCode, body)
	}
	if string(body) != wantResponse {
		t.Errorf("raw hostcall fixture response = %q, want %q", body, wantResponse)
	}
}

func writeHostCallConfig(t *testing.T, fixturePath string) string {
	t.Helper()

	config := map[string]any{
		"services": map[string]any{
			"hostcall-contracts": map[string]any{
				"name": "hostcall-contracts",
				"functions": map[string]any{
					"kvlogging": map[string]any{
						"filepath": fixturePath,
					},
				},
				"routes": []map[string]any{
					{
						"type":     "http",
						"path":     "/kvlogging",
						"methods":  []string{"GET"},
						"function": "kvlogging",
					},
				},
			},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal raw hostcall fixture config: %v", err)
	}
	path := filepath.Join(t.TempDir(), "tarmac.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write raw hostcall fixture config: %v", err)
	}
	return path
}

func waitForHostCallServer(client *http.Client, healthURL string, runErr <-chan error) error {
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-runErr:
			if err == nil {
				return errors.New("hostcall compatibility server stopped without an error")
			}
			if errors.Is(err, ErrShutdown) {
				return errors.New("hostcall compatibility server stopped before becoming ready")
			}
			return fmt.Errorf("run hostcall compatibility server: %w", err)
		case <-deadline.C:
			return errors.New("hostcall compatibility server did not become ready")
		case <-ticker.C:
			response, err := client.Get(healthURL)
			if err != nil {
				continue
			}
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}
