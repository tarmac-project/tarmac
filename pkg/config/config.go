/*
Package config provides a parser for a JSON configuration file format used to configure Tarmac functions and services.
*/
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Config is a struct that represents the parsed configuration file. It contains a map of Tarmac Services, where each
// Service is identified by a unique string key.
type Config struct {
	sync.RWMutex

	// routes is an internal mapping of routes and functions.
	routes map[string]string

	// Services maps the names of Tarmac services to their configurations, which include the set of functions they provide
	// and the routes by which they can be invoked.
	Services map[string]Service `json:"services"`
}

// Service defines a Tarmac service, which consists of a Name, a set of Functions, and a collection of available Routes.
type Service struct {
	// Name is the human-readable name for the service.
	Name string `json:"name"`

	// Functions is a map of Function objects, where the keys are string identifiers for each function.
	Functions map[string]Function `json:"functions"`

	// Routes is a slice of Route objects representing the available routes for this service.
	Routes []Route `json:"routes"`
}

// Function defines the Tarmac function to load and execute.
type Function struct {
	// Filepath to the WASM function
	Filepath string `json:"filepath"`
}

// Route defines available routes for the service.
type Route struct {
	// Type defines the protocol used for the route, examples are "http", "nats", "scheduled_task", etc.
	Type string `json:"type"`

	// Path defines the path for HTTP types.
	Path string `json:"path,omitempty"`

	// Topic defines the topic or channel to listen to for message queue based routes.
	Topic string `json:"topic,omitempty"`

	// Methods defines the HTTP methods to accept for the defined route.
	Methods []string `json:"methods,omitempty"`

	// Function defines the Function to execute when this route is called.
	Function string `json:"function"`

	// Frequency is used to define the frequency of scheduled_task routes in seconds.
	Frequency int `json:"frequency,omitempty"`
}

// ErrRouteNotFound is returned when a requested route is not found in the
// service's route configuration.
var ErrRouteNotFound = fmt.Errorf("route not found")

// Parse function reads the file specified and attempts to parse the contents into a Config instance.
func Parse(filepath string) (*Config, error) {
	// Read the file contents
	b, err := os.ReadFile(filepath)
	if err == os.ErrNotExist {
		return &Config{}, os.ErrNotExist
	}
	if err != nil {
		return &Config{}, fmt.Errorf("could not read service configuration: %w", err)
	}

	// Unmarshal the JSON contents into a Config struct
	cfg := &Config{}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return &Config{}, fmt.Errorf("could not parse service configuration file: %w", err)
	}

	// Populate the internal routes map
	cfg.routes = make(map[string]string)
	for _, svcCfg := range cfg.Services {
		for _, r := range svcCfg.Routes {
			if r.Type == "http" {
				for _, m := range r.Methods {
					key := fmt.Sprintf("%s:%s:%s", r.Type, m, r.Path)
					cfg.routes[key] = r.Function
				}
			}
		}
	}

	return cfg, nil
}

// RouteLookup searches the routes map for a given key and returns the corresponding function name if the key is found,
//
//	or an empty string and an error if it is not found.
func (cfg *Config) RouteLookup(key string) (string, error) {
	cfg.RLock()
	defer cfg.RUnlock()
	v, ok := cfg.routes[key]
	if !ok {
		return "", ErrRouteNotFound
	}
	return v, nil
}
