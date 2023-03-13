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
	"github.com/madflojo/hord/drivers/cassandra"
	"github.com/madflojo/hord/drivers/redis"
	"github.com/madflojo/tarmac/pkg/callbacks"
	"github.com/madflojo/tarmac/pkg/callbacks/httpclient"
	"github.com/madflojo/tarmac/pkg/callbacks/kvstore"
	"github.com/madflojo/tarmac/pkg/callbacks/logging"
	"github.com/madflojo/tarmac/pkg/callbacks/metrics"
	sqlstore "github.com/madflojo/tarmac/pkg/callbacks/sql"
	"github.com/madflojo/tarmac/pkg/config"
	"github.com/madflojo/tarmac/pkg/telemetry"
	"github.com/madflojo/tarmac/pkg/tlsconfig"
	"github.com/madflojo/tarmac/pkg/wasm"
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

// srv is the global reference for the HTTP Server.
var srv *server

// engine is the global WASM Engine
var engine *wasm.Server

// kv is the global reference for the K/V Store.
var kv hord.Database

// db is the global reference for the SQL DB.
var db *sql.DB

// runCtx is a global context used to control shutdown of the application.
var runCtx context.Context

// runCancel is a global context cancelFunc used to trigger the shutdown of applications.
var runCancel context.CancelFunc

// cfg is used across the app package to contain configuration.
var cfg *viper.Viper

// log is used across the app package for logging.
var log *logrus.Logger

// scheduler is a internal task scheduler for recurring tasks
var scheduler *tasks.Scheduler

// stats is used across the app package to manage and access system metrics.
var stats = telemetry.New()

// Run starts the primary application. It handles starting background services,
// populating package globals & structures, and clean up tasks.
func Run(c *viper.Viper) error {
	var err error

	// Create App Context
	runCtx, runCancel = context.WithCancel(context.Background())

	// Apply config provided by main to the package global
	cfg = c

	// Initiate a new logger
	log = logrus.New()
	if cfg.GetBool("debug") {
		log.Level = logrus.DebugLevel
		log.Debug("Enabling Debug Logging")
	}
	if cfg.GetBool("trace") {
		log.Level = logrus.TraceLevel
		log.Debug("Enabling Trace Logging")
	}
	if cfg.GetBool("disable_logging") {
		log.Level = logrus.FatalLevel
	}

	// Setup Scheduler
	scheduler = tasks.New()
	defer scheduler.Stop()

	// Config Reload
	if cfg.GetInt("config_watch_interval") > 0 && cfg.GetBool("use_consul") {
		_, err := scheduler.Add(&tasks.Task{
			Interval: time.Duration(cfg.GetInt("config_watch_interval")) * time.Second,
			TaskFunc: func() error {
				// Reload config using Viper's Watch capabilities
				err := cfg.WatchRemoteConfig()
				if err != nil {
					return err
				}

				// Support hot enable/disable of debug logging
				if cfg.GetBool("debug") {
					log.Level = logrus.DebugLevel
				}

				// Support hot enable/disable of trace logging
				if cfg.GetBool("trace") {
					log.Level = logrus.TraceLevel
				}

				// Support hot enable/disable of all logging
				if cfg.GetBool("disable_logging") {
					log.Level = logrus.FatalLevel
				}

				log.Tracef("Config reloaded from Consul")
				return nil
			},
		})
		if err != nil {
			log.Errorf("Error scheduling Config watcher - %s", err)
		}
	}

	// Setup the KV Connection
	if cfg.GetBool("enable_kvstore") {
		log.Infof("Connecting to KV Store")
		switch cfg.GetString("kvstore_type") {
		case "redis":
			kv, err = redis.Dial(redis.Config{
				Server:   cfg.GetString("redis_server"),
				Password: cfg.GetString("redis_password"),
				SentinelConfig: redis.SentinelConfig{
					Servers: cfg.GetStringSlice("redis_sentinel_servers"),
					Master:  cfg.GetString("redis_sentinel_master"),
				},
				ConnectTimeout: time.Duration(cfg.GetInt("redis_connect_timeout")) * time.Second,
				Database:       cfg.GetInt("redis_database"),
				SkipTLSVerify:  cfg.GetBool("redis_hostname_verify"),
				KeepAlive:      time.Duration(cfg.GetInt("redis_keepalive")) * time.Second,
				MaxActive:      cfg.GetInt("redis_max_active"),
				ReadTimeout:    time.Duration(cfg.GetInt("redis_read_timeout")) * time.Second,
				WriteTimeout:   time.Duration(cfg.GetInt("redis_write_timeout")) * time.Second,
			})
			if err != nil {
				return fmt.Errorf("could not establish kvstore connection - %s", err)
			}
		case "cassandra":
			kv, err = cassandra.Dial(cassandra.Config{
				Hosts:                      cfg.GetStringSlice("cassandra_hosts"),
				Port:                       cfg.GetInt("cassandra_port"),
				Keyspace:                   cfg.GetString("cassandra_keyspace"),
				Consistency:                cfg.GetString("cassandra_consistency"),
				ReplicationStrategy:        cfg.GetString("cassandra_repl_strategy"),
				Replicas:                   cfg.GetInt("cassandra_replicas"),
				User:                       cfg.GetString("cassandra_user"),
				Password:                   cfg.GetString("cassandra_password"),
				EnableHostnameVerification: cfg.GetBool("cassandra_hostname_verify"),
			})
			if err != nil {
				return fmt.Errorf("could not establish kvstore connection - %s", err)
			}
		default:
			return fmt.Errorf("unknown kvstore specified - %s", cfg.GetString("kvstore_type"))
		}

		// Clean up KV Store connections on shutdown
		defer kv.Close()

		// Initialize the KV
		err = kv.Setup()
		if err != nil {
			return fmt.Errorf("could not setup kvstore - %s", err)
		}
	}

	if kv == nil {
		log.Infof("KV Store not configured, skipping")
	}

	if cfg.GetBool("enable_sql") {
		log.Infof("Connecting to SQL DB")
		switch cfg.GetString("sql_type") {
		case "mysql":
			db, err = sql.Open("mysql", cfg.GetString("sql_dsn"))
			if err != nil {
				return fmt.Errorf("could not establish sql db connection - %s", err)
			}
		case "postgres":
			db, err = sql.Open("postgres", cfg.GetString("sql_dsn"))
			if err != nil {
				return fmt.Errorf("could not establish sql db connection - %s", err)
			}
		default:
			return fmt.Errorf("unknown sql store specified - %s", cfg.GetString("sql_type"))
		}
	}
	if db == nil {
		log.Infof("SQL DB not configured, skipping")
	}

	// Setup the HTTP Server
	srv = &server{
		httpRouter: httprouter.New(),
	}
	srv.httpServer = &http.Server{
		Addr:    cfg.GetString("listen_addr"),
		Handler: srv.httpRouter,
	}

	// Setup TLS Configuration
	if cfg.GetBool("enable_tls") {
		tlsCfg := tlsconfig.New()

		// Load Certs from file
		err := tlsCfg.CertsFromFile(cfg.GetString("cert_file"), cfg.GetString("key_file"))
		if err != nil {
			return fmt.Errorf("unable to configure HTTPS server with certificate and key - %s", err)
		}

		// Load CA enabling m-TLS
		if cfg.GetString("ca_file") != "" {
			err := tlsCfg.CAFromFile(cfg.GetString("ca_file"))
			if err != nil {
				return fmt.Errorf("unable to configure HTTPS server with provided client certificate authority - %s", err)
			}

			// Set to ask but ignore client certs
			if cfg.GetBool("ignore_client_cert") {
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
		log.Infof("Received shutdown signal %s", s)

		defer Stop()
	}()

	// Register Health Check Handler used for Liveness checks
	srv.httpRouter.GET("/health", srv.middleware(srv.Health))

	// Register Health Check Handler used for Readiness checks
	srv.httpRouter.GET("/ready", srv.middleware(srv.Ready))

	// Create WASM Callback Router
	router := callbacks.New(callbacks.Config{
		PreFunc: func(namespace, op string, data []byte) ([]byte, error) {
			log.WithFields(logrus.Fields{
				"namespace": namespace,
				"function":  op,
			}).Debugf("CallbackRouter called with payload %s", data)
			return []byte(""), nil
		},
		PostFunc: func(r callbacks.CallbackResult) {
			// Measure Callback Execution time and counts
			stats.Callbacks.WithLabelValues(fmt.Sprintf("%s:%s", r.Namespace, r.Operation)).Observe(r.EndTime.Sub(r.StartTime).Seconds())

			// Log Callback failures as warnings
			if r.Err != nil {
				log.WithFields(logrus.Fields{
					"namespace": r.Namespace,
					"function":  r.Operation,
				}).Warnf("Callback call resulted in error after %f seconds - %s", r.EndTime.Sub(r.StartTime).Seconds(), r.Err)
			}
		},
	})

	// Setup SQL Callbacks
	if cfg.GetBool("enable_sql") {
		cbSQL, err := sqlstore.New(sqlstore.Config{DB: db})
		if err != nil {
			return fmt.Errorf("unable to initialize callback sqlstore for WASM functions - %s", err)
		}

		// Register SQLStore Callbacks
		router.RegisterCallback("sql", "query", cbSQL.Query)
	}

	// Setup KVStore Callbacks
	if cfg.GetBool("enable_kvstore") {
		cbKVStore, err := kvstore.New(kvstore.Config{KV: kv})
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
		Log: log,
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
	engine, err = wasm.NewServer(wasm.Config{
		Callback: router.Callback,
	})
	if err != nil {
		return err
	}

	// Look for Functions Config
	funcCfg, err := config.Parse(cfg.GetString("wasm_function_config"))
	if err != nil {
		log.Infof("Could not load wasm_function_config (%s) starting with default function path - %s", cfg.GetString("wasm_function_config"), err)

		// Load WASM Function using default path
		err = engine.LoadModule(wasm.ModuleConfig{
			Name:     "default",
			Filepath: cfg.GetString("wasm_function"),
		})
		if err != nil {
			return fmt.Errorf("could not load default function path for wasm_function (%s) - %s", cfg.GetString("wasm_function"), err)
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
		log.Infof("Loading Services from wasm_function_config %s", cfg.GetString("wasm_function_config"))

		for svcName, svcCfg := range funcCfg.Services {
			// Load WASM Functions
			log.Infof("Loading Functions from Service %s", svcName)
			for fName, fCfg := range svcCfg.Functions {
				err := engine.LoadModule(wasm.ModuleConfig{
					Name:     fName,
					Filepath: fCfg.Filepath,
				})
				if err != nil {
					return fmt.Errorf("could not load function %s from path %s - %s", fName, fCfg.Filepath, err)
				}
				log.Infof("Loaded Function %s for Service %s", fName, svcName)
			}

			// Register Routes
			log.Infof("Registering Routes from Service %s", svcName)
			funcRoutes := make(map[string]string)
			for _, r := range svcCfg.Routes {
				if r.Type == "http" {
					for _, m := range r.Methods {
						key := fmt.Sprintf("%s:%s:%s", r.Type, m, r.Path)
						log.Infof("Registering Route %s for function %s", key, r.Function)
						funcRoutes[key] = r.Function
						srv.httpRouter.Handle(m, r.Path, srv.middleware(srv.WASMHandler))
					}
				}

				if r.Type == "scheduled_task" {
					id, err := scheduler.Add(&tasks.Task{
						Interval: time.Duration(r.Frequency) * time.Second,
						TaskFunc: func() error {
							now := time.Now()
							_, err := runWASM(r.Function, "handler", []byte(""))
							if err != nil {
								stats.Tasks.WithLabelValues(r.Function).Observe(time.Since(now).Seconds())
								return err
							}
							stats.Tasks.WithLabelValues(r.Function).Observe(time.Since(now).Seconds())
							return nil
						},
					})
					if err != nil {
						log.Errorf("Error scheduling scheduled task %s - %s", r.Function, err)
					}

					// Clean up Task on Shutdown
					defer scheduler.Del(id)
				}

			}
			srv.funcRoutes = funcRoutes
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
	log.Infof("Starting HTTP Listener on %s", cfg.GetString("listen_addr"))
	if cfg.GetBool("enable_tls") {
		err := srv.httpServer.ListenAndServeTLS(cfg.GetString("cert_file"), cfg.GetString("key_file"))
		if err != nil {
			if err == http.ErrServerClosed {
				// Wait until all outstanding requests are done
				<-runCtx.Done()
				return ErrShutdown
			}
			return fmt.Errorf("unable to start HTTPS Server - %s", err)
		}
	}
	err = srv.httpServer.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			// Wait until all outstanding requests are done
			<-runCtx.Done()
			return ErrShutdown
		}
		return fmt.Errorf("unable to start HTTP Server - %s", err)
	}

	return nil
}

// Stop is used to gracefully shutdown the server.
func Stop() {
	err := srv.httpServer.Shutdown(context.Background())
	if err != nil {
		log.Errorf("Unexpected error while shutting down HTTP server - %s", err)
	}
	defer runCancel()
}
