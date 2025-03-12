/*
Package app is the primary runtime service.
*/
package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	pprof "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	// MySQL Database Driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"

	// PostgreSQL Database Driver
	_ "github.com/lib/pq"
	"github.com/madflojo/tasks"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/tarmac-project/hord"
	"github.com/tarmac-project/hord/drivers/bbolt"
	"github.com/tarmac-project/hord/drivers/cassandra"
	"github.com/tarmac-project/hord/drivers/hashmap"
	"github.com/tarmac-project/hord/drivers/redis"
	"github.com/tarmac-project/tarmac/pkg/callbacks/httpclient"
	"github.com/tarmac-project/tarmac/pkg/callbacks/kvstore"
	"github.com/tarmac-project/tarmac/pkg/callbacks/logging"
	"github.com/tarmac-project/tarmac/pkg/callbacks/metrics"
	sqlstore "github.com/tarmac-project/tarmac/pkg/callbacks/sql"
	"github.com/tarmac-project/tarmac/pkg/config"
	"github.com/tarmac-project/tarmac/pkg/telemetry"
	"github.com/tarmac-project/tarmac/pkg/tlsconfig"
	"github.com/tarmac-project/wapc-toolkit/callbacks"
	"github.com/tarmac-project/wapc-toolkit/engine"
)

// Common errors returned by this app.
var (
	ErrShutdown = fmt.Errorf("application shutdown gracefully")
)

// LevelNames maps custom log levels to their string representations
var LevelNames = map[slog.Leveler]string{
	LevelTrace:      "TRACE",
	slog.LevelDebug: "DEBUG",
	slog.LevelInfo:  "INFO",
	slog.LevelWarn:  "WARN",
	slog.LevelError: "ERROR",
}

const (
	// DefaultNamespace is the default namespace for callback functions.
	DefaultNamespace = "tarmac"

	// RouteTypeInit is the route type for init functions.
	RouteTypeInit = "init"

	// RouteTypeHTTP is the route type for HTTP functions.
	RouteTypeHTTP = "http"

	// RouteTypeScheduledTask is the route type for scheduled tasks.
	RouteTypeScheduledTask = "scheduled_task"

	// RouteTypeFunction is the route type for function to function calls.
	RouteTypeFunction = "function"

	// LevelTrace is a custom log level for trace logging.
	LevelTrace = slog.LevelDebug - 4

	// LevelDisabled is a custom log level for disabled logging.
	LevelDisabled = slog.LevelError + 4
)

// Server represents the main server structure.
type Server struct {
	// cfg is used across the app package to contain configuration.
	cfg *viper.Viper

	// db is the global reference for the SQL DB.
	db *sql.DB

	// engine is the global WASM Engine.
	engine *engine.Server

	// funcCfg is used to store and access multi-function service configurations.
	funcCfg *config.Config

	// httpRouter is used to store and access the HTTP Request Router.
	httpRouter *httprouter.Router

	// httpServer is the primary HTTP server.
	httpServer *http.Server

	// kv is the global reference for the K/V Store.
	kv hord.Database

	// log is used across the app package for logging.
	log *slog.Logger

	// logLeveler is used to dynamically change the log level.
	logLeveler *slog.LevelVar

	// runCancel is a global context cancelFunc used to trigger the shutdown of applications.
	runCancel context.CancelFunc

	// runCtx is a global context used to control shutdown of the application.
	runCtx context.Context

	// scheduler is an internal task scheduler for recurring tasks.
	scheduler *tasks.Scheduler

	// stats is used across the app package to manage and access system metrics.
	stats *telemetry.Telemetry
}

// New creates a new instance of the Server struct.
// It takes a `cfg` parameter of type `*viper.Viper` for configuration.
// It returns a pointer to the created Server instance.
func New(cfg *viper.Viper) *Server {
	srv := &Server{cfg: cfg}

	// Create App Context
	srv.runCtx, srv.runCancel = context.WithCancel(context.Background())

	// Create a dynamic level variable for runtime log level changes
	// This allows us to change log levels without recreating handlers
	srv.logLeveler = new(slog.LevelVar)

	// Set initial log level
	srv.logLeveler.Set(slog.LevelInfo)
	if srv.cfg.GetBool("debug") {
		srv.logLeveler.Set(slog.LevelDebug)
	}

	if srv.cfg.GetBool("trace") {
		srv.logLeveler.Set(LevelTrace)
	}

	if srv.cfg.GetBool("disable_logging") {
		srv.logLeveler.Set(LevelDisabled)
	}

	// Create handler options with our dynamic level var and custom level names
	handlerOpts := &slog.HandlerOptions{
		Level: srv.logLeveler,
		// Replace the level attribute to use our custom level names
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				if k, ok := a.Value.Any().(slog.Level); ok {
					levelLabel, exists := LevelNames[k]
					if exists {
						a.Value = slog.StringValue(levelLabel)
					}
					a.Value = slog.StringValue(levelLabel)
				}
			}

			return a
		},
	}

	// Create a JSON or Text handler based on config
	var handler slog.Handler
	handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	if srv.cfg.GetBool("text_log_format") {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	// Initialize the logger
	srv.log = slog.New(handler)

	// Log the log level settings
	if srv.cfg.GetBool("debug") {
		srv.log.Debug("Enabling Debug Logging")
	}

	if srv.cfg.GetBool("trace") {
		srv.log.Debug("Enabling Trace Logging")
	}

	return srv
}

// Run starts the primary application. It handles starting background services,
// populating package globals & structures, and clean up tasks.
func (srv *Server) Run() error {
	var err error

	// Setup Stats
	srv.stats = telemetry.New()
	defer srv.stats.Close()

	// Setup Scheduler
	srv.scheduler = tasks.New()
	defer srv.scheduler.Stop()

	// Startup Log Message
	srv.log.Info("Starting Tarmac",
		"run_mode", srv.cfg.GetString("run_mode"),
		"use_consul", srv.cfg.GetBool("use_consul"),
		"enable_kvstore", srv.cfg.GetBool("enable_kvstore"),
		"enable_sql", srv.cfg.GetBool("enable_sql"),
		"enable_tls", srv.cfg.GetBool("enable_tls"),
		"enable_metrics", srv.cfg.GetBool("enable_metrics"),
	)

	// Config Reload
	if srv.cfg.GetInt("config_watch_interval") > 0 && srv.cfg.GetBool("use_consul") {
		_, err := srv.scheduler.Add(&tasks.Task{
			Interval: time.Duration(srv.cfg.GetInt("config_watch_interval")) * time.Second,
			TaskFunc: func() error {
				// Reload config using Viper's Watch capabilities
				err := srv.cfg.WatchRemoteConfig()
				if err != nil {
					return err
				}

				// Start with default info level
				newLevel := slog.LevelInfo

				// Set appropriate level based on config
				if srv.cfg.GetBool("debug") {
					newLevel = slog.LevelDebug
				}

				if srv.cfg.GetBool("trace") {
					// Use our custom trace level
					newLevel = LevelTrace
				}

				if srv.cfg.GetBool("disable_logging") {
					newLevel = slog.LevelError + 4
				}

				// Update the level
				if newLevel != srv.logLeveler.Level() {
					srv.logLeveler.Set(newLevel)

					// Log the change
					if srv.cfg.GetBool("debug") {
						srv.log.Debug("Dynamic log level updated to debug")
					} else if srv.cfg.GetBool("trace") {
						srv.log.Debug("Dynamic log level updated to trace")
					}
				}

				srv.log.Log(context.Background(), LevelTrace, "Config Reloaded from Consul")
				return nil
			},
		})
		if err != nil {
			srv.log.Error("Error scheduling Config watcher: "+err.Error(), "error", err)
		}
	}

	// Setup the KV Connection
	if srv.cfg.GetBool("enable_kvstore") {
		srv.log.Info("Connecting to KV Store")
		switch srv.cfg.GetString("kvstore_type") {
		case "in-memory":
			srv.kv, err = hashmap.Dial(hashmap.Config{})
			if err != nil {
				return fmt.Errorf("could not create internal kvstore - %w", err)
			}
		case "internal", "boltdb":
			// Check if file exists, if not create one
			fh, err := os.OpenFile(srv.cfg.GetString("boltdb_filename"), os.O_RDWR|os.O_CREATE|os.O_EXCL, os.FileMode(srv.cfg.GetInt("boltdb_permissions")))
			if err != nil && !os.IsExist(err) {
				return fmt.Errorf("could not create boltdb file - %w", err)
			}
			fh.Close()

			// Open datastore
			srv.kv, err = bbolt.Dial(bbolt.Config{
				Filename:    srv.cfg.GetString("boltdb_filename"),
				Bucketname:  srv.cfg.GetString("boltdb_bucket"),
				Permissions: os.FileMode(srv.cfg.GetInt("boltdb_permissions")),
				Timeout:     time.Duration(srv.cfg.GetInt("boltdb_timeout")) * time.Second,
			})
			if err != nil {
				return fmt.Errorf("could not create internal kvstore - %w", err)
			}
		case "redis":
			srv.kv, err = redis.Dial(redis.Config{
				Server:   srv.cfg.GetString("redis_server"),
				Password: srv.cfg.GetString("redis_password"),
				SentinelConfig: redis.SentinelConfig{
					Servers: srv.cfg.GetStringSlice("redis_sentinel_servers"),
					Master:  srv.cfg.GetString("redis_sentinel_master"),
				},
				ConnectTimeout: time.Duration(srv.cfg.GetInt("redis_connect_timeout")) * time.Second,
				Database:       srv.cfg.GetInt("redis_database"),
				SkipTLSVerify:  srv.cfg.GetBool("redis_hostname_verify"),
				KeepAlive:      time.Duration(srv.cfg.GetInt("redis_keepalive")) * time.Second,
				MaxActive:      srv.cfg.GetInt("redis_max_active"),
				ReadTimeout:    time.Duration(srv.cfg.GetInt("redis_read_timeout")) * time.Second,
				WriteTimeout:   time.Duration(srv.cfg.GetInt("redis_write_timeout")) * time.Second,
			})
			if err != nil {
				return fmt.Errorf("could not establish kvstore connection - %s", err)
			}
		case "cassandra":
			srv.kv, err = cassandra.Dial(cassandra.Config{
				Hosts:                      srv.cfg.GetStringSlice("cassandra_hosts"),
				Port:                       srv.cfg.GetInt("cassandra_port"),
				Keyspace:                   srv.cfg.GetString("cassandra_keyspace"),
				Consistency:                srv.cfg.GetString("cassandra_consistency"),
				ReplicationStrategy:        srv.cfg.GetString("cassandra_repl_strategy"),
				Replicas:                   srv.cfg.GetInt("cassandra_replicas"),
				User:                       srv.cfg.GetString("cassandra_user"),
				Password:                   srv.cfg.GetString("cassandra_password"),
				EnableHostnameVerification: srv.cfg.GetBool("cassandra_hostname_verify"),
			})
			if err != nil {
				return fmt.Errorf("could not establish kvstore connection - %s", err)
			}
		default:
			return fmt.Errorf("unknown kvstore specified - %s", srv.cfg.GetString("kvstore_type"))
		}

		// Clean up KV Store connections on shutdown
		defer srv.kv.Close()

		// Initialize the KV
		err = srv.kv.Setup()
		if err != nil {
			return fmt.Errorf("could not setup kvstore - %s", err)
		}
	}

	if srv.kv == nil {
		srv.log.Info("KV Store not configured, skipping")
	}

	if srv.cfg.GetBool("enable_sql") {
		srv.log.Info("Connecting to SQL DB")
		switch srv.cfg.GetString("sql_type") {
		case "mysql":
			srv.db, err = sql.Open("mysql", srv.cfg.GetString("sql_dsn"))
			if err != nil {
				return fmt.Errorf("could not establish sql db connection - %s", err)
			}
		case "postgres":
			srv.db, err = sql.Open("postgres", srv.cfg.GetString("sql_dsn"))
			if err != nil {
				return fmt.Errorf("could not establish sql db connection - %s", err)
			}
		default:
			return fmt.Errorf("unknown sql store specified - %s", srv.cfg.GetString("sql_type"))
		}
	}
	if srv.db == nil {
		srv.log.Info("SQL DB not configured, skipping")
	}

	// Setup the HTTP Server
	srv.httpRouter = httprouter.New()
	srv.httpServer = &http.Server{
		Addr:    srv.cfg.GetString("listen_addr"),
		Handler: srv.httpRouter,
	}

	// Setup TLS Configuration
	if srv.cfg.GetBool("enable_tls") {
		tlsCfg := tlsconfig.New()

		// Load Certs from file
		err := tlsCfg.CertsFromFile(srv.cfg.GetString("cert_file"), srv.cfg.GetString("key_file"))
		if err != nil {
			return fmt.Errorf("unable to configure HTTPS server with certificate and key - %s", err)
		}

		// Load CA enabling m-TLS
		if srv.cfg.GetString("ca_file") != "" {
			err := tlsCfg.CAFromFile(srv.cfg.GetString("ca_file"))
			if err != nil {
				return fmt.Errorf("unable to configure HTTPS server with provided client certificate authority - %s", err)
			}

			// Set to ask but ignore client certs
			if srv.cfg.GetBool("ignore_client_cert") {
				tlsCfg.IgnoreClientCert()
			}
		}

		// Generate TLS config and assign to HTTP Server
		srv.httpServer.TLSConfig = tlsCfg.Generate()
	}

	// Kick off Graceful Shutdown Go Routine
	go func() {
		// Make the Trap
		trap := make(chan os.Signal, 1)
		signal.Notify(trap, syscall.SIGTERM)

		// Wait for a signal then action
		s := <-trap
		srv.log.Info("Received shutdown signal", "signal", s)

		defer srv.Stop()
	}()

	// Register Health Check Handler used for Liveness checks
	srv.httpRouter.GET("/health", srv.middleware(srv.Health))

	// Register Health Check Handler used for Readiness checks
	srv.httpRouter.GET("/ready", srv.middleware(srv.Ready))

	// Create WASM Callback Router
	router, err := callbacks.New(callbacks.RouterConfig{
		PreFunc: func(rq callbacks.CallbackRequest) ([]byte, error) {
			// Debug logging of callback
			srv.log.Debug("CallbackRouter called",
				"namespace", rq.Namespace,
				"capability", rq.Capability,
				"operation", rq.Operation,
				"callback_start_time", rq.StartTime.String(),
			)

			// Trace logging of callback
			srv.log.Log(context.Background(), LevelTrace, "CallbackRouter called with payload",
				"namespace", rq.Namespace,
				"capability", rq.Capability,
				"operation", rq.Operation,
				"callback_start_time", rq.StartTime.String(),
				"payload", string(rq.Input),
			)
			return []byte(""), nil
		},
		PostFunc: func(r callbacks.CallbackResult) {
			// Measure Callback Execution time and counts
			duration := r.EndTime.Sub(r.StartTime).Milliseconds()
			srv.stats.Callbacks.WithLabelValues(fmt.Sprintf("%s:%s", r.Namespace, r.Operation)).Observe(float64(duration))

			// Debug logging of callback results
			if r.Err != nil {
				srv.log.Debug("Callback returned result with error: "+r.Err.Error(),
					"namespace", r.Namespace,
					"capability", r.Capability,
					"operation", r.Operation,
					"error", r.Err,
					"duration", duration,
					"duration_ms", duration,
				)
			} else {
				srv.log.Debug("Callback returned result successfully",
					"namespace", r.Namespace,
					"capability", r.Capability,
					"operation", r.Operation,
					"duration", duration,
					"duration_ms", duration,
				)
			}

			// Trace logging of callback results
			if r.Err != nil {
				srv.log.Log(context.Background(), LevelTrace, "Callback returned result with error and output: "+r.Err.Error(),
					"namespace", r.Namespace,
					"capability", r.Capability,
					"operation", r.Operation,
					"error", r.Err,
					"input", string(r.Input),
					"duration", duration,
					"duration_ms", duration,
					"output", string(r.Output),
				)
			} else {
				srv.log.Log(context.Background(), LevelTrace, "Callback returned result with output",
					"namespace", r.Namespace,
					"capability", r.Capability,
					"operation", r.Operation,
					"input", string(r.Input),
					"duration", duration,
					"duration_ms", duration,
					"output", string(r.Output),
				)
			}

			// Log Callback failures as warnings
			if r.Err != nil {
				srv.log.Warn("Callback call resulted in error: "+r.Err.Error(),
					"namespace", r.Namespace,
					"capability", r.Capability,
					"operation", r.Operation,
					"duration", duration,
					"duration_ms", duration,
					"error", r.Err,
				)
			}
		},
	})
	if err != nil {
		return fmt.Errorf("unable to initialize callback router - %s", err)
	}

	// Start WASM Engine
	srv.engine, err = engine.New(engine.ServerConfig{
		Callback: router.Callback,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize wasm engine - %s", err)
	}

	// Setup SQL Callbacks
	if srv.cfg.GetBool("enable_sql") {
		cbSQL, err := sqlstore.New(sqlstore.Config{DB: srv.db})
		if err != nil {
			return fmt.Errorf("unable to initialize callback sqlstore for WASM functions - %s", err)
		}

		// Register SQLStore Callbacks
		err = router.RegisterCallback(callbacks.CallbackConfig{
			Namespace:  DefaultNamespace,
			Capability: "sql",
			Operation:  "query",
			Func:       cbSQL.Query,
		})
		if err != nil {
			return fmt.Errorf("unable to register callback for sql query - %s", err)
		}
	}

	// Setup KVStore Callbacks
	if srv.cfg.GetBool("enable_kvstore") {
		cbKVStore, err := kvstore.New(kvstore.Config{KV: srv.kv})
		if err != nil {
			return fmt.Errorf("unable to initialize callback kvstore for WASM functions - %s", err)
		}

		// Register KVStore Callbacks
		err = router.RegisterCallback(callbacks.CallbackConfig{
			Namespace:  DefaultNamespace,
			Capability: "kvstore",
			Operation:  "get",
			Func:       cbKVStore.Get,
		})
		if err != nil {
			return fmt.Errorf("unable to register callback for kvstore get - %s", err)
		}

		err = router.RegisterCallback(callbacks.CallbackConfig{
			Namespace:  DefaultNamespace,
			Capability: "kvstore",
			Operation:  "set",
			Func:       cbKVStore.Set,
		})
		if err != nil {
			return fmt.Errorf("unable to register callback for kvstore set - %s", err)
		}

		err = router.RegisterCallback(callbacks.CallbackConfig{
			Namespace:  DefaultNamespace,
			Capability: "kvstore",
			Operation:  "delete",
			Func:       cbKVStore.Delete,
		})
		if err != nil {
			return fmt.Errorf("unable to register callback for kvstore delete - %s", err)
		}

		err = router.RegisterCallback(callbacks.CallbackConfig{
			Namespace:  DefaultNamespace,
			Capability: "kvstore",
			Operation:  "keys",
			Func:       cbKVStore.Keys,
		})
		if err != nil {
			return fmt.Errorf("unable to register callback for kvstore keys - %s", err)
		}
	}

	// Setup HTTP Callbacks
	cbHTTPClient, err := httpclient.New(httpclient.Config{})
	if err != nil {
		return fmt.Errorf("unable to initialize callback http client for WASM functions - %s", err)
	}

	// Register HTTPClient Functions
	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "httpclient",
		Operation:  "call",
		Func:       cbHTTPClient.Call,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for httpclient call - %s", err)
	}

	// Setup Logger Callbacks
	cbLogger, err := logging.New(logging.Config{
		// Pass general logger into host callback with adapter
		Log: logging.NewSlogAdapter(srv.log),
	})
	if err != nil {
		return fmt.Errorf("unable to initialize callback logger for WASM functions - %s", err)
	}

	// Register Logger Functions
	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "logger",
		Operation:  "info",
		Func:       cbLogger.Info,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for logger info - %s", err)
	}

	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "logger",
		Operation:  "error",
		Func:       cbLogger.Error,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for logger error - %s", err)
	}

	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "logger",
		Operation:  "warn",
		Func:       cbLogger.Warn,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for logger warn - %s", err)
	}

	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "logger",
		Operation:  "debug",
		Func:       cbLogger.Debug,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for logger debug - %s", err)
	}

	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "logger",
		Operation:  "trace",
		Func:       cbLogger.Trace,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for logger trace - %s", err)
	}

	// Setup Metrics Callbacks
	cbMetrics, err := metrics.New(metrics.Config{})
	if err != nil {
		return fmt.Errorf("unable to initialize callback metrics for WASM functions - %s", err)
	}

	// Register Metrics Callbacks
	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "metrics",
		Operation:  "counter",
		Func:       cbMetrics.Counter,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for metrics counter - %s", err)
	}

	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "metrics",
		Operation:  "gauge",
		Func:       cbMetrics.Gauge,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for metrics gauge - %s", err)
	}

	err = router.RegisterCallback(callbacks.CallbackConfig{
		Namespace:  DefaultNamespace,
		Capability: "metrics",
		Operation:  "histogram",
		Func:       cbMetrics.Histogram,
	})
	if err != nil {
		return fmt.Errorf("unable to register callback for metrics histogram - %s", err)
	}

	// Look for Functions Config
	srv.funcCfg, err = config.Parse(srv.cfg.GetString("wasm_function_config"))
	if err != nil {
		srv.log.Info("Could not load wasm_function_config starting with default function path",
			"config_path", srv.cfg.GetString("wasm_function_config"),
			"error", err)

		// Load WASM Function using default path
		err = srv.engine.LoadModule(engine.ModuleConfig{
			Name:     "default",
			Filepath: srv.cfg.GetString("wasm_function"),
			PoolSize: srv.cfg.GetInt("wasm_pool_size"),
		})
		if err != nil {
			return fmt.Errorf("could not load default function path for wasm_function (%s) - %s", srv.cfg.GetString("wasm_function"), err)
		}

		// Register WASM Handler with default path
		srv.httpRouter.GET("/", srv.middleware(srv.WASMHandler))
		srv.httpRouter.POST("/", srv.middleware(srv.WASMHandler))
		srv.httpRouter.PUT("/", srv.middleware(srv.WASMHandler))
		srv.httpRouter.DELETE("/", srv.middleware(srv.WASMHandler))
		srv.httpRouter.HEAD("/", srv.middleware(srv.WASMHandler))

		// Measure Routes
		srv.stats.Routes.WithLabelValues("default", "http").Inc()
	}

	// Load Functions from Config
	if err == nil {
		srv.log.Info("Loading Services from wasm_function_config", "config_path", srv.cfg.GetString("wasm_function_config"))

		routesCounter := map[string]int{
			RouteTypeInit:          0,
			RouteTypeHTTP:          0,
			RouteTypeScheduledTask: 0,
			RouteTypeFunction:      0,
		}
		for svcName, svcCfg := range srv.funcCfg.Services {
			// Load WASM Functions
			srv.log.Info("Loading Functions from Service", "service", svcName)
			for fName, fCfg := range svcCfg.Functions {
				err := srv.engine.LoadModule(engine.ModuleConfig{
					Name:     fName,
					Filepath: fCfg.Filepath,
					PoolSize: fCfg.PoolSize,
				})
				if err != nil {
					return fmt.Errorf("could not load function %s from path %s - %s", fName, fCfg.Filepath, err)
				}
				srv.log.Info("Loaded Function for Service",
					"function", fName,
					"service", svcName,
					"filepath", fCfg.Filepath)
			}

			// Register Routes
			srv.log.Info("Registering Routes from Service",
				"service", svcName)
			funcRoutes := make(map[string]string)
			initRoutes := []config.Route{}
			for _, r := range svcCfg.Routes {
				switch r.Type {
				case RouteTypeInit:
					// Copy init functions for later execution
					initRoutes = append(initRoutes, r)
					routesCounter[RouteTypeInit]++
					srv.stats.Routes.WithLabelValues(svcName, r.Type).Inc()

				case RouteTypeHTTP:
					// Register HTTP based functions with the HTTP router
					for _, m := range r.Methods {
						key := fmt.Sprintf("%s:%s:%s", r.Type, m, r.Path)
						srv.log.Info("Registering Route for function",
							"function", r.Function,
							"method", m,
							"path", r.Path,
							"function_type", r.Type,
							"service", svcName,
							"route_key", key)
						funcRoutes[key] = r.Function
						srv.httpRouter.Handle(m, r.Path, srv.middleware(srv.WASMHandler))
						routesCounter[RouteTypeHTTP]++
						srv.stats.Routes.WithLabelValues(svcName, r.Type).Inc()
					}

				case RouteTypeScheduledTask:
					// Schedule tasks for scheduled functions
					fname := r.Function
					srv.log.Info("Scheduling custom task for function",
						"function", r.Function,
						"interval", r.Frequency)
					id, err := srv.scheduler.Add(&tasks.Task{
						Interval: time.Duration(r.Frequency) * time.Second,
						TaskFunc: func() error {
							now := time.Now()
							srv.log.Log(context.Background(), LevelTrace, "Executing Scheduled Task", "function", fname)
							_, err := srv.runWASM(fname, "handler", []byte(""))
							if err != nil {
								srv.stats.Tasks.WithLabelValues(fname).Observe(float64(time.Since(now).Milliseconds()))
								return err
							}
							srv.stats.Tasks.WithLabelValues(fname).Observe(float64(time.Since(now).Milliseconds()))
							return nil
						},
					})
					if err != nil {
						srv.log.Error("Error scheduling scheduled task: "+err.Error(), "function", r.Function, "error", err)
					}
					// Clean up Task on Shutdown
					defer srv.scheduler.Del(id)
					routesCounter[RouteTypeScheduledTask]++
					srv.stats.Routes.WithLabelValues(svcName, r.Type).Inc()

				case RouteTypeFunction:
					// Setup callbacks for function to function calls
					srv.log.Info("Registering Function to Function callback", "function", r.Function)
					fname := r.Function
					f := func(b []byte) ([]byte, error) {
						srv.log.Info("Executing Function to Function callback", "function", fname)
						return srv.runWASM(fname, "handler", b)
					}
					err := router.RegisterCallback(callbacks.CallbackConfig{
						Namespace:  DefaultNamespace,
						Capability: "function",
						Operation:  fname,
						Func:       f,
					})
					if err != nil {
						return fmt.Errorf("error registering callback for function %s - %s", fname, err)
					}
					routesCounter[RouteTypeFunction]++
					srv.stats.Routes.WithLabelValues(svcName, r.Type).Inc()
				}
			}

			// Execute init functions
			for _, r := range initRoutes {
				srv.log.Info("Executing Init Function", "function", r.Function)
				var success, retries int
				delay := r.Frequency
				for success == 0 && retries <= r.Retries {
					// Execute the function
					_, err := srv.runWASM(r.Function, "handler", []byte(""))
					if err != nil {
						srv.log.Error("Error executing Init Function: "+err.Error(),
							"function", r.Function,
							"error", err)
						retries++
						// Wait exponentially longer between retries
						<-time.After(time.Duration(delay) * time.Second)
						delay *= delay
						continue
					}
					success = 1
				}
				if success == 0 {
					return fmt.Errorf("init function %s exceeded retries", r.Function)
				}
			}

		}

		// Log information about loaded functions and routes
		srv.log.Info("Loaded Functions and Routes",
			"init", routesCounter[RouteTypeInit],
			"http", routesCounter[RouteTypeHTTP],
			"scheduled_task", routesCounter[RouteTypeScheduledTask],
			"function", routesCounter[RouteTypeFunction])

		// If run-mode is jobs, exit cleanly
		if srv.cfg.GetString("run_mode") == "job" {
			srv.log.Info("Run mode is job, exiting after init function execution")
			return ErrShutdown
		}
	}

	// Register Metrics Handler
	srv.httpRouter.GET("/metrics", srv.handlerWrapper(promhttp.Handler()))

	// Register PProf Handlers
	srv.httpRouter.GET("/debug/pprof/", srv.handlerWrapper(http.HandlerFunc(pprof.Index)))
	srv.httpRouter.GET("/debug/pprof/cmdline", srv.handlerWrapper(http.HandlerFunc(pprof.Cmdline)))
	srv.httpRouter.GET("/debug/pprof/profile", srv.handlerWrapper(http.HandlerFunc(pprof.Profile)))
	srv.httpRouter.GET("/debug/pprof/symbol", srv.handlerWrapper(http.HandlerFunc(pprof.Symbol)))
	srv.httpRouter.GET("/debug/pprof/trace", srv.handlerWrapper(http.HandlerFunc(pprof.Trace)))
	srv.httpRouter.GET("/debug/pprof/allocs", srv.handlerWrapper(pprof.Handler("allocs")))
	srv.httpRouter.GET("/debug/pprof/mutex", srv.handlerWrapper(pprof.Handler("mutex")))
	srv.httpRouter.GET("/debug/pprof/goroutine", srv.handlerWrapper(pprof.Handler("goroutine")))
	srv.httpRouter.GET("/debug/pprof/heap", srv.handlerWrapper(pprof.Handler("heap")))
	srv.httpRouter.GET("/debug/pprof/threadcreate", srv.handlerWrapper(pprof.Handler("threadcreate")))
	srv.httpRouter.GET("/debug/pprof/block", srv.handlerWrapper(pprof.Handler("block")))

	// Start HTTP Listener
	srv.log.Info("Starting HTTP Listener", "address", srv.cfg.GetString("listen_addr"))
	if srv.cfg.GetBool("enable_tls") {
		err := srv.httpServer.ListenAndServeTLS(srv.cfg.GetString("cert_file"), srv.cfg.GetString("key_file"))
		if err != nil {
			if err == http.ErrServerClosed {
				// Wait until all outstanding requests are done
				<-srv.runCtx.Done()
				return ErrShutdown
			}
			return fmt.Errorf("unable to start HTTPS Server - %s", err)
		}
	}
	err = srv.httpServer.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			// Wait until all outstanding requests are done
			<-srv.runCtx.Done()
			return ErrShutdown
		}
		return fmt.Errorf("unable to start HTTP Server - %s", err)
	}

	return nil
}

// Stop is used to gracefully shutdown the server.
func (srv *Server) Stop() {
	srv.stats.Close()
	err := srv.httpServer.Shutdown(context.Background())
	if err != nil {
		srv.log.Error("Unexpected error while shutting down HTTP server: "+err.Error(), "error", err)
	}
	defer srv.runCancel()
}
