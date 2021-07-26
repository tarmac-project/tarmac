/*
Package app is the primary runtime service.
*/
package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/madflojo/hord"
	"github.com/madflojo/hord/drivers/cassandra"
	"github.com/madflojo/hord/drivers/redis"
	"github.com/madflojo/tarmac"
	"github.com/madflojo/tarmac/callbacks"
	"github.com/madflojo/tarmac/wasm"
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
var stats = newMetrics()

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

	// Setup the HTTP Server
	srv = &server{
		httpRouter: httprouter.New(),
		kvStore:    &kvStore{},
		logger:     &logger{},
	}
	srv.httpServer = &http.Server{
		Addr:    cfg.GetString("listen_addr"),
		Handler: srv.httpRouter,
	}

	// Setup TLS Configuration
	if cfg.GetBool("enable_tls") {
		srv.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
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
			stats.callbacks.WithLabelValues(fmt.Sprintf("%s:%s", namespace, op)).Inc()
			log.WithFields(logrus.Fields{
				"namespace": namespace,
				"function":  op,
			}).Infof("CallbackRouter called with payload %s", data)
			return []byte(""), nil
		},
	})

	// Setup KVStore Callbacks
	if cfg.GetBool("enable_kvstore") {
		router.RegisterCallback("kvstore", "get", srv.kvStore.Get)
		router.RegisterCallback("kvstore", "set", srv.kvStore.Set)
		router.RegisterCallback("kvstore", "delete", srv.kvStore.Delete)
		router.RegisterCallback("kvstore", "keys", srv.kvStore.Keys)
	}

	// Setup Logger Callbacks
	router.RegisterCallback("logger", "info", srv.logger.Info)
	router.RegisterCallback("logger", "error", srv.logger.Error)
	router.RegisterCallback("logger", "warn", srv.logger.Warn)
	router.RegisterCallback("logger", "debug", srv.logger.Debug)
	router.RegisterCallback("logger", "trace", srv.logger.Trace)

	// Start WASM Engine
	engine, err = wasm.NewServer(wasm.Config{
		Callback: router.Callback,
	})
	if err != nil {
		return err
	}

	// Preload Default WASM Function
	if cfg.GetString("wasm_function") != "" {
		err = engine.LoadModule(wasm.ModuleConfig{
			Name:     "default",
			Filepath: cfg.GetString("wasm_function"),
		})
		if err != nil {
			return err
		}
	}

	// Setup Scheduler based tasks
	for k := range cfg.GetStringMap("scheduled_tasks") {
		name := k
		log.Infof("Scheduling Task - %s", name)

		// Preload WASM Function
		err = engine.LoadModule(wasm.ModuleConfig{
			Name:     "scheduler-" + name,
			Filepath: cfg.GetString("scheduled_tasks." + name + ".wasm_function"),
		})
		if err != nil {
			log.Errorf("Error loading WASM module for scheduled task %s - %s", name, err)
		}

		// Create Scheduled Task
		headers := cfg.GetStringMapString("scheduled_tasks." + name + ".headers")
		headers["request_type"] = "scheduler"
		id, err := scheduler.Add(&tasks.Task{
			Interval: time.Duration(cfg.GetInt("scheduled_tasks."+name+".interval")) * time.Second,
			TaskFunc: func() error {
				now := time.Now()
				log.WithFields(logrus.Fields{"task-name": name}).Tracef("Executing Scheduled task")
				r, err := runWASM("scheduler-"+name, "scheduler:RUN", tarmac.ServerRequest{Headers: headers})
				if err != nil {
					log.WithFields(logrus.Fields{"task-name": name}).Debugf("Error executing task - %s", err)
					stats.tasks.WithLabelValues(name).Observe(time.Since(now).Seconds())
					return err
				}
				if r.Status.Code == 200 {
					log.WithFields(logrus.Fields{"task-name": name}).Debugf("Task execution completed successfully")
				}
				stats.tasks.WithLabelValues(name).Observe(time.Since(now).Seconds())
				return nil
			},
		})
		if err != nil {
			log.Errorf("Error scheduling scheduled task %s - %s", name, err)
		}

		// Clean up Task on Shutdown
		defer scheduler.Del(id)
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

	// Register WASM Handler
	srv.httpRouter.GET("/", srv.middleware(srv.WASMHandler))
	srv.httpRouter.POST("/", srv.middleware(srv.WASMHandler))
	srv.httpRouter.PUT("/", srv.middleware(srv.WASMHandler))
	srv.httpRouter.DELETE("/", srv.middleware(srv.WASMHandler))
	srv.httpRouter.HEAD("/", srv.middleware(srv.WASMHandler))

	// Start HTTP Listener
	log.Infof("Starting Listener on %s", cfg.GetString("listen_addr"))
	if cfg.GetBool("enable_tls") {
		log.Infof("Using Certificate: %s Key: %s", cfg.GetString("cert_file"), cfg.GetString("key_file"))
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
