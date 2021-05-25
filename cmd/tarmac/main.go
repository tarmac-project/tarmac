package main

import (
	"github.com/madflojo/tarmac/app"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	// Initiate a simple logger
	log := logrus.New()

	// Setup Config
	cfg := viper.New()

	// Set Default Configs
	cfg.SetDefault("enable_tls", true)
	cfg.SetDefault("listen_addr", "0.0.0.0:8443")
	cfg.SetDefault("config_watch_interval", 15)
	cfg.SetDefault("wasm_module", "./tarmac_module.wasm")

	// Load Config
	cfg.AddConfigPath("./conf")
	cfg.SetEnvPrefix("app")
	cfg.AllowEmptyEnv(true)
	cfg.AutomaticEnv()
	err := cfg.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Warnf("No Config file found, loaded config from Environment - Default path ./conf")
		default:
			log.Fatalf("Error when Fetching Configuration - %s", err)
		}
	}

	// Load Config from Consul
	if cfg.GetBool("use_consul") {
		log.Infof("Setting up Consul Config source - %s/%s", cfg.GetString("consul_addr"), cfg.GetString("consul_keys_prefix"))
		err = cfg.AddRemoteProvider("consul", cfg.GetString("consul_addr"), cfg.GetString("consul_keys_prefix"))
		if err != nil {
			log.Fatalf("Error adding Consul as a remote Configuration Provider - %s", err)
		}
		cfg.SetConfigType("json")
		err = cfg.ReadRemoteConfig()
		if err != nil {
			log.Fatalf("Error when Fetching Configuration from Consul - %s", err)
		}

		if cfg.GetBool("from_consul") {
			log.Infof("Successfully loaded configuration from consul")
		}
	}

	// Run application
	err = app.Run(cfg)
	if err != nil && err != app.ErrShutdown {
		log.Fatalf("Service stopped - %s", err)
	}
	log.Infof("Service shutdown - %s", err)
}
