/*
Package app is the primary runtime service.
*/
package app

import (
	"context"
	"database/sql"
	"fmt"
	// MySQL Database Driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	// PostgreSQL Database Driver
	_ "github.com/lib/pq"
	"github.com/madflojo/hord"
	"github.com/madflojo/hord/drivers/bbolt"
	"github.com/madflojo/hord/drivers/cassandra"
	"github.com/madflojo/hord/drivers/hashmap"
	"github.com/madflojo/hord/drivers/redis"
	"github.com/tarmac-project/tarmac/pkg/callbacks"
	"github.com/tarmac-project/tarmac/pkg/callbacks/httpclient"
	"github.com/tarmac-project/tarmac/pkg/callbacks/kvstore"
	"github.com/tarmac-project/tarmac/pkg/callbacks/logging"
	"github.com/tarmac-project/tarmac/pkg/callbacks/metrics"
	sqlstore "github.com/tarmac-project/tarmac/pkg/callbacks/sql"
	"github.com/tarmac-project/tarmac/pkg/config"
	"github.com/tarmac-project/tarmac/pkg/telemetry"
	"github.com/tarmac-project/tarmac/pkg/tlsconfig"
	"github.com/tarmac-project/tarmac/pkg/wasm"
	"github.com/madflojo/tasks"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	pprof "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Common errors returned by this app.
var (
	ErrShutdown = fmt.Errorf("application shutdown gracefully")
)

// Server represents the main server structure.
type Server struct {
	// cfg is used across the app package to contain configuration.
	cfg *viper.Viper

	// db is the global reference for the SQL DB.
	db *sql.DB

	// engine is the global WASM Engine.
	engine *wasm.Server

	// funcCfg is used to store and access multi-function service configurations.
	funcCfg *config.Config

	// httpRouter is used to store and access the HTTP Request Router.
	httpRouter *httprouter.Router

	// httpServer is the primary HTTP server.
	httpServer *http.Server

	// kv is the global reference for the K/V Store.
	kv hord.Database

	// log is used across the app package for logging.
	log *logrus.Logger

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

	// Initiate a new logger
	srv.log = logrus.New()
	if srv.cfg.GetBool("debug") {
		srv.log.Level = logrus.DebugLevel
		srv.log.Debug("Enabling Debug Logging")
	}
	if srv.cfg.GetBool("trace") {
		srv.log.Level = logrus.TraceLevel
		srv.log.Debug("Enabling Trace Logging")
	}
	if srv.cfg.GetBool("disable_logging") {
		srv.log.Level = logrus.FatalLevel
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

				// Support hot enable/disable of debug logging
				if srv.cfg.GetBool("debug") {
					srv.log.Level = logrus.DebugLevel
				}

				// Support hot enable/disable of trace logging
				if srv.cfg.GetBool("trace") {
					srv.log.Level = logrus.TraceLevel
				}

				// Support hot enable/disable of all logging
				if srv.cfg.GetBool("disable_logging") {
					srv.log.Level = logrus.FatalLevel
				}

				srv.log.Tracef("Config reloaded from Consul")
				return nil
			},
		})
		if err != nil {
			srv.log.Errorf("Error scheduling Config watcher - %s", err)
		}
	}

	// Setup the KV Connection
	if srv.cfg.GetBool("enable_kvstore") {
		srv.log.Infof("Connecting to KV Store")
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
		srv.log.Infof("KV Store not configured, skipping")
	}

	if srv.cfg.GetBool("enable_sql") {
		srv.log.Infof("Connecting to SQL DB")
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
		srv.log.Infof("SQL DB not configured, skipping")
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
		srv.log.Infof("Received shutdown signal %s", s)

		defer srv.Stop()
	}()

	// Register Health Check Handler used for Liveness checks
	srv.httpRouter.GET("/health", srv.middleware(srv.Health))

	// Register Health Check Handler used for Readiness checks
	srv.httpRouter.GET("/ready", srv.middleware(srv.Ready))

	// Create WASM Callback Router
	router := callbacks.New(callbacks.Config{
		PreFunc: func(namespace, op string, data []byte) ([]byte, error) {
			// Trace logging of callback
			srv.log.WithFields(logrus.Fields{
				"namespace": namespace,
				"function":  op,
			}).Tracef("CallbackRouter called with payload %s", data)
			return []byte(""), nil
		},
		PostFunc: func(r callbacks.CallbackResult) {
			// Measure Callback Execution time and counts
			srv.stats.Callbacks.WithLabelValues(fmt.Sprintf("%s:%s", r.Namespace, r.Operation)).Observe(r.EndTime.Sub(r.StartTime).Seconds())

			// Trace logging of callback results
			srv.log.WithFields(logrus.Fields{
				"namespace": r.Namespace,
				"function":  r.Operation,
				"input":     r.Input,
				"error":     r.Err,
			}).Tracef("Callback returned result after %f seconds with output - %s", r.EndTime.Sub(r.StartTime).Seconds(), r.Output)

			// Log Callback failures as warnings
			if r.Err != nil {
				srv.log.WithFields(logrus.Fields{
					"namespace": r.Namespace,
					"function":  r.Operation,
				}).Warnf("Callback call resulted in error after %f seconds - %s", r.EndTime.Sub(r.StartTime).Seconds(), r.Err)
			}
		},
	})

	// Setup SQL Callbacks
	if srv.cfg.GetBool("enable_sql") {
		cbSQL, err := sqlstore.New(sqlstore.Config{DB: srv.db})
		if err != nil {
			return fmt.Errorf("unable to initialize callback sqlstore for WASM functions - %s", err)
		}

		// Register SQLStore Callbacks
		router.RegisterCallback("sql", "query", cbSQL.Query)
	}

	// Setup KVStore Callbacks
	if srv.cfg.GetBool("enable_kvstore") {
		cbKVStore, err := kvstore.New(kvstore.Config{KV: srv.kv})
		if err != nil {
			return fmt.Errorf("unable to initialize callback kvstore for WASM functions - %s", err)
		}

		// Register KVStore Callbacks
		router.RegisterCallback("kvstore", "get", cbKVStore.Get)
		router.RegisterCallback("kvstore", "set", cbKVStore.Set)
		router.RegisterCallback("kvstore", "delete", cbKVStore.Delete)
		router.RegisterCallback("kvstore", "keys", cbKVStore.Keys)
	}

	// Setup HTTP Callbacks
	cbHTTPClient, err := httpclient.New(httpclient.Config{})
	if err != nil {
		return fmt.Errorf("unable to initialize callback http client for WASM functions - %s", err)
	}

	// Register HTTPClient Functions
	router.RegisterCallback("httpclient", "call", cbHTTPClient.Call)

	// Setup Logger Callbacks
	cbLogger, err := logging.New(logging.Config{
		// Pass general logger into host callback
		Log: srv.log,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize callback logger for WASM functions - %s", err)
	}

	// Register Logger Functions
	router.RegisterCallback("logger", "info", cbLogger.Info)
	router.RegisterCallback("logger", "error", cbLogger.Error)
	router.RegisterCallback("logger", "warn", cbLogger.Warn)
	router.RegisterCallback("logger", "debug", cbLogger.Debug)
	router.RegisterCallback("logger", "trace", cbLogger.Trace)

	// Setup Metrics Callbacks
	cbMetrics, err := metrics.New(metrics.Config{})
	if err != nil {
		return fmt.Errorf("unable to initialize callback metrics for WASM functions - %s", err)
	}

	// Register Metrics Callbacks
	router.RegisterCallback("metrics", "counter", cbMetrics.Counter)
	router.RegisterCallback("metrics", "gauge", cbMetrics.Gauge)
	router.RegisterCallback("metrics", "histogram", cbMetrics.Histogram)

	// Start WASM Engine
	srv.engine, err = wasm.NewServer(wasm.Config{
		Callback: router.Callback,
	})
	if err != nil {
		return err
	}

	// Look for Functions Config
	srv.funcCfg, err = config.Parse(srv.cfg.GetString("wasm_function_config"))
	if err != nil {
		srv.log.Infof("Could not load wasm_function_config (%s) starting with default function path - %s", srv.cfg.GetString("wasm_function_config"), err)

		// Load WASM Function using default path
		err = srv.engine.LoadModule(wasm.ModuleConfig{
			Name:     "default",
			Filepath: srv.cfg.GetString("wasm_function"),
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
	}

	// Load Functions from Config
	if err == nil {
		srv.log.Infof("Loading Services from wasm_function_config %s", srv.cfg.GetString("wasm_function_config"))

		for svcName, svcCfg := range srv.funcCfg.Services {
			// Load WASM Functions
			srv.log.Infof("Loading Functions from Service %s", svcName)
			for fName, fCfg := range svcCfg.Functions {
				err := srv.engine.LoadModule(wasm.ModuleConfig{
					Name:     fName,
					Filepath: fCfg.Filepath,
				})
				if err != nil {
					return fmt.Errorf("could not load function %s from path %s - %s", fName, fCfg.Filepath, err)
				}
				srv.log.Infof("Loaded Function %s for Service %s", fName, svcName)
			}

			// Register Routes
			srv.log.Infof("Registering Routes from Service %s", svcName)
			funcRoutes := make(map[string]string)
			for _, r := range svcCfg.Routes {
				if r.Type == "http" {
					for _, m := range r.Methods {
						key := fmt.Sprintf("%s:%s:%s", r.Type, m, r.Path)
						srv.log.Infof("Registering Route %s for function %s", key, r.Function)
						funcRoutes[key] = r.Function
						srv.httpRouter.Handle(m, r.Path, srv.middleware(srv.WASMHandler))
					}
				}

				if r.Type == "scheduled_task" {
					srv.log.Infof("Scheduling custom task for function %s with interval of %d", r.Function, r.Frequency)
					id, err := srv.scheduler.Add(&tasks.Task{
						Interval: time.Duration(r.Frequency) * time.Second,
						TaskFunc: func() error {
							now := time.Now()
							srv.log.Tracef("Executing Scheduled Function %s", r.Function)
							_, err := srv.runWASM(r.Function, "handler", []byte(""))
							if err != nil {
								srv.stats.Tasks.WithLabelValues(r.Function).Observe(time.Since(now).Seconds())
								return err
							}
							srv.stats.Tasks.WithLabelValues(r.Function).Observe(time.Since(now).Seconds())
							return nil
						},
					})
					if err != nil {
						srv.log.Errorf("Error scheduling scheduled task %s - %s", r.Function, err)
					}

					// Clean up Task on Shutdown
					defer srv.scheduler.Del(id)
				}

				if r.Type == "function" {
					srv.log.Infof("Registering Function to Function callback for %s", r.Function)
					router.RegisterCallback("function", r.Function, func(b []byte) ([]byte, error) {
						return srv.runWASM(r.Function, "handler", b)
					})
				}
			}
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
	srv.log.Infof("Starting HTTP Listener on %s", srv.cfg.GetString("listen_addr"))
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
		srv.log.Errorf("Unexpected error while shutting down HTTP server - %s", err)
	}
	defer srv.runCancel()
}
