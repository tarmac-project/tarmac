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

	// PoolSize defines the number of instances of the function to create
	PoolSize int `json:"pool_size"`
}

// Route defines available routes for the service.
type Route struct {
	// Type defines the Route type for the function.
	//
	// Valid types are:
	// - http - HTTP based routes
	// - function - Function to Function calls
	// - scheduled_task - Scheduled function calls
	// - init - Initialization functions
	Type string `json:"type"`

	// Path defines the path for HTTP types.
	Path string `json:"path,omitempty"`

	// Topic defines the topic or channel to listen to for message queue based routes.
	Topic string `json:"topic,omitempty"`

	// Methods defines the HTTP methods to accept for the defined route.
	Methods []string `json:"methods,omitempty"`

	// Function defines the Function to execute when this route is called.
	Function string `json:"function"`

	// Frequency is used for both scheduled_task and init routes, with the value being the interval in seconds.
	//
	// When defined as an scheduled_task route, frequency is the interval at which tasks are executed.
	// If no frequency is defined, an error will occur.
	//
	// When defined as an init route, frequency is used to define the interval in seconds between retries.
	// As the number of retries increases, the interval will exponentially increase.
	// If no frequency is defined, the default value is 1 second.
	Frequency int `json:"frequency,omitempty"`

	// Retries is used to define the number of retries for init routes.
	// If the init route fails, it will be retried for the number of times defined with a exponential backoff.
	// The default value is 0 which means no retries.
	Retries int `json:"retries,omitempty"`
}

var (
	// ErrRouteNotFound is returned when a requested route is not found in the
	// service's route configuration.
	ErrRouteNotFound = fmt.Errorf("route not found")

	// ErrInvalidConfig is returned when the configuration file does not contain the required fields or is otherwise
	// invalid.
	ErrInvalidConfig = fmt.Errorf("invalid configuration file")
)

const (
	// DefaultPoolSize is the default number of instances of a function to create.
	DefaultPoolSize = 100

	// DefaultFrequency is the default frequency for init routes.
	DefaultFrequency = 1
)

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

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return &Config{}, ErrInvalidConfig
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

// Validate function checks the configuration for required fields and returns an error if any are missing.
// Validate will also populate any default values for fields that are not defined.
func (cfg *Config) Validate() error {
	cfg.Lock()
	defer cfg.Unlock()

	// Loop through each service and validate the configuration
	for sk, svcCfg := range cfg.Services {
		// Validate the service name
		if svcCfg.Name == "" {
			return fmt.Errorf("service missing name: %w", ErrInvalidConfig)
		}

		// Validate functions
		for fk, f := range svcCfg.Functions {
			if f.Filepath == "" {
				return fmt.Errorf("function missing filepath: %w", ErrInvalidConfig)
			}
			if f.PoolSize == 0 {
				f.PoolSize = DefaultPoolSize
				cfg.Services[sk].Functions[fk] = f
			}
		}

		// Validate routes
		for rk, r := range svcCfg.Routes {
			if r.Type == "" || r.Function == "" {
				return fmt.Errorf("route missing type or function: %w", ErrInvalidConfig)
			}

			// Validate the route type
			switch r.Type {
			case "http":
				if r.Path == "" || len(r.Methods) == 0 {
					return fmt.Errorf("http route missing path or methods: %w", ErrInvalidConfig)
				}
			case "scheduled_task":
				if r.Frequency == 0 {
					return fmt.Errorf("scheduled_task route missing frequency: %w", ErrInvalidConfig)
				}
			case "init":
				if r.Frequency == 0 {
					cfg.Services[sk].Routes[rk].Frequency = DefaultFrequency
				}
			}
		}
	}
	return nil
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
