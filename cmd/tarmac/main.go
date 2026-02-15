package main

import (
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"github.com/tarmac-project/tarmac/pkg/app"
	"log/slog"
	"os"
)

func main() {
	// Initiate a simple logger
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Setup Config
	cfg := viper.New()

	// Set Default Configs
	cfg.SetDefault("enable_tls", true)
	cfg.SetDefault("listen_addr", "0.0.0.0:8443")
	cfg.SetDefault("cert_file", "/certs/cert.crt")
	cfg.SetDefault("key_file", "/certs/key.key")
	cfg.SetDefault("config_watch_interval", 15)
	cfg.SetDefault("wasm_function", "/functions/tarmac.wasm")
	cfg.SetDefault("wasm_function_config", "/functions/tarmac.json")
	cfg.SetDefault("kvstore_type", "internal")
	cfg.SetDefault("boltdb_filename", "/data/tarmac/tarmac.db")
	cfg.SetDefault("boltdb_bucket", "tarmac")
	cfg.SetDefault("boltdb_permissions", 0600)
	cfg.SetDefault("boltdb_timeout", 5)
	cfg.SetDefault("nats_url", "nats://localhost:4222")
	cfg.SetDefault("nats_bucket", "tarmac")
	cfg.SetDefault("grpc_socket_path", "/grpc.sock")
	cfg.SetDefault("run_mode", "daemon")
	cfg.SetDefault("text_log_format", false)
	cfg.SetDefault("http_client_max_response_body_size", 10*1024*1024) // 10MB default

	// Load Config
	cfg.AddConfigPath("./conf")
	cfg.SetEnvPrefix("app")
	cfg.AllowEmptyEnv(true)
	cfg.AutomaticEnv()
	err := cfg.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Warn("No Config file found, loaded config from Environment - Default path ./conf")
		default:
			log.Error("Error when Fetching Configuration: "+err.Error(), "error", err)
			os.Exit(1)
		}
	}

	// Load Config from Consul
	if cfg.GetBool("use_consul") {
		log.Info("Setting up Consul Config source",
			"consul_addr", cfg.GetString("consul_addr"),
			"consul_keys_prefix", cfg.GetString("consul_keys_prefix"))
		err = cfg.AddRemoteProvider("consul", cfg.GetString("consul_addr"), cfg.GetString("consul_keys_prefix"))
		if err != nil {
			log.Error("Error adding Consul as a remote Configuration Provider: "+err.Error(), "error", err)
			os.Exit(1)
		}
		cfg.SetConfigType("json")
		err = cfg.ReadRemoteConfig()
		if err != nil {
			log.Error("Error when Fetching Configuration from Consul: "+err.Error(), "error", err)
			os.Exit(1)
		}

		if cfg.GetBool("from_consul") {
			log.Info("Successfully loaded configuration from consul")
		}
	}

	// Run application
	srv := app.New(cfg)
	defer srv.Stop()
	err = srv.Run()
	if err != nil && err != app.ErrShutdown {
		log.Error("Service stopped: "+err.Error(), "error", err)
		os.Exit(1)
	}
	log.Info("Service shutdown", "error", err)
}
