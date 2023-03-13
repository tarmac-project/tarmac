/*
Package config provides a parser for a JSON configuration file format used to configure Tarmac functions and services.
*/
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is a struct that represents the parsed configuration file. It contains a map of Tarmac Services, where each
// Service is identified by a unique string key.
type Config struct {
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

	// Frequency is used to define the frequency of scheduled_task routes. Example frequencies are `1s`, `2m`, `3h`,
	// and `4d`.
	Frequency string `json:"frequency,omitempty"`
}

// Parse function reads the file specified and attempts to parse the contents into a Config instance.
func Parse(filepath string) (Config, error) {
	// Read the file contents
	b, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, fmt.Errorf("could not read service configuration: %w", err)
	}

	// Unmarshal the JSON contents into a Config struct
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("could not parse service configuration file: %w", err)
	}

	return cfg, nil
}
