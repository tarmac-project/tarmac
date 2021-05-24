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
	"github.com/madflojo/hord/drivers/redis"
	"github.com/madflojo/tasks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
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
var engine *WASMServer

// db is the global reference for the DB Server.
var db hord.Database

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
	if cfg.GetInt("config_watch_interval") > 0 {
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

	// Setup the DB Connection
	if cfg.GetString("db_server") != "" {
		db, err = redis.Dial(redis.Config{
			Server:   cfg.GetString("db_server"),
			Password: cfg.GetString("db_password"),
		})
		if err != nil {
			return fmt.Errorf("could not establish database connection - %s", err)
		}
		defer db.Close()

		// Initialize the DB
		err = db.Setup()
		if err != nil {
			return fmt.Errorf("could not setup database - %s", err)
		}
	}

	if db == nil {
		log.Infof("Database not configured, skipping")
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

		// Shutdown the HTTP Server
		err := srv.httpServer.Shutdown(context.Background())
		if err != nil {
			log.Errorf("Received errors when shutting down HTTP sessions %s", err)
		}

		// Close DB Sessions
		db.Close()

		// Shutdown the app via runCtx
		runCancel()
	}()

	// Register Health Check Handler used for Liveness checks
	srv.httpRouter.GET("/health", srv.middleware(srv.Health))

	// Register Health Check Handler used for Readiness checks
	srv.httpRouter.GET("/ready", srv.middleware(srv.Ready))

	// Start WASM Engine
	engine, err = NewWASMServer()
	if err != nil {
		return err
	}

	if cfg.GetString("wasm_module") != "" {
		err = engine.LoadModule("http_handler", cfg.GetString("wasm_module"))
		if err != nil {
			return err
		}
	}

	// Register WASM Handler
	srv.httpRouter.GET("/", srv.middleware(srv.WASMHandler))
	srv.httpRouter.POST("/", srv.middleware(srv.WASMHandler))
	srv.httpRouter.PUT("/", srv.middleware(srv.WASMHandler))
	srv.httpRouter.DELETE("/", srv.middleware(srv.WASMHandler))

	// Start HTTP Listener
	log.Infof("Starting Listener on %s", cfg.GetString("listen_addr"))
	if cfg.GetBool("enable_tls") {
		err := srv.httpServer.ListenAndServeTLS(cfg.GetString("cert_file"), cfg.GetString("key_file"))
		if err != nil {
			if err == http.ErrServerClosed {
				// Wait until all outstanding requests are done
				<-runCtx.Done()
				return ErrShutdown
			}
			return err
		}
	}
	err = srv.httpServer.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			// Wait until all outstanding requests are done
			<-runCtx.Done()
			return ErrShutdown
		}
		return err
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
