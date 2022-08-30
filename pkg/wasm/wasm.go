/*
Package wasm is a Web Assembly Runtime wrapper for Tarmac.

This package provides the ability to start a WASM engine server, load modules, and invoke functions within those WASM modules.
*/
package wasm

import (
	"context"
	"fmt"
	"github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
	"os"
	"sync"
	"time"
)

const (
	// Default Pool Size.
	DefaultPoolSize = 100

	// Default Pool Timeout.
	DefaultPoolTimeout = 5
)

// Config is used to configure the initial WASM Server.
type Config struct {

	// Callback is a function provided by the caller, this callback function is used when WASM modules perform a host callback.
	Callback func(context.Context, string, string, string, []byte) ([]byte, error)
}

// Server is used as a WASM runtime engine providing capabilities to load and run modules.
type Server struct {
	sync.RWMutex

	// callback is provided by the caller, this callback function is used when WASM modules perform a host callback.
	callback func(context.Context, string, string, string, []byte) ([]byte, error)

	// modules is a map for storing and fetching modules that have already been loaded.
	modules map[string]*Module
}

// ModuleConfig is used to configure a specific WASM module for the Server to load and ready for execution.
type ModuleConfig struct {
	// Name is the name of this module, this is used as a lookup key when serving many modules.
	Name string

	// Filepath is the file path to read the module from.
	Filepath string

	// PoolSize is used to control the size of the WASM Module Pool.
	PoolSize int
}

// Module is a specific WASM Module that can be loaded into the engine.
type Module struct {
	// Name is the name of the WASM module.
	Name string

	// ctx is a context used to clean up module instances
	ctx context.Context

	// cancel is a context cancellation function
	cancel context.CancelFunc

	// module is the loaded module, this is referenced for clean up and closure purposes.
	module wapc.Module

	// pool is the module pool created as part of loading a module. This pool is used to store and fetch module instances as needed.
	pool *wapc.Pool

	// poolSize will determine the size of a module pool.
	poolSize uint64
}

// NewServer will create a new Server with the Engine and Module Store pre-loaded.
func NewServer(cfg Config) (*Server, error) {
	s := &Server{}
	s.modules = make(map[string]*Module)

	if cfg.Callback == nil {
		return s, fmt.Errorf("Callback cannot be nil")
	}

	s.callback = cfg.Callback
	return s, nil
}

// Shutdown will shutdown the WASM Server cleaning up any open modules, pools, or instances.
func (s *Server) Shutdown() {
	s.RLock()
	defer s.RUnlock()
	for _, m := range s.modules {
		defer m.cancel()
		defer m.module.Close(m.ctx)
		defer m.pool.Close(m.ctx)
	}
}

// LoadModule will read and load the WASM module from the filesysem.
func (s *Server) LoadModule(cfg ModuleConfig) error {
	if cfg.Name == "" || cfg.Filepath == "" {
		return fmt.Errorf("key and file cannot be empty")
	}

	// Create Module
	m := &Module{
		Name: cfg.Name,
	}

	// Create context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Set Pool Size
	m.poolSize = uint64(cfg.PoolSize)
	if cfg.PoolSize == 0 {
		m.poolSize = uint64(DefaultPoolSize)
	}

	// Read the WASM module file
	guest, err := os.ReadFile(cfg.Filepath)
	if err != nil {
		return fmt.Errorf("unable to read wasm module file - %s", err)
	}

	// Initiate waPC Engine
	engine := wazero.Engine()

	// Create a new Module from file contents
	m.module, err = engine.New(m.ctx, s.callback, guest, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return fmt.Errorf("unable to load module with wasm file %s - %s", cfg.Filepath, err)
	}

	// Create pool for module
	m.pool, err = wapc.NewPool(m.ctx, m.module, m.poolSize)
	if err != nil {
		return fmt.Errorf("unable to create module pool for wasm file %s - %s", cfg.Filepath, err)
	}

	s.Lock()
	defer s.Unlock()
	s.modules[m.Name] = m

	return nil
}

// Module will return the WASMModule stored for the specified WASM module.
func (s *Server) Module(key string) (*Module, error) {
	var m *Module
	s.RLock()
	defer s.RUnlock()
	if m, ok := s.modules[key]; ok {
		return m, nil
	}
	return m, fmt.Errorf("module not found")
}

// Run will fetch an instance from the module pool and execute it.
func (m *Module) Run(handler string, payload []byte) ([]byte, error) {
	var r []byte
	i, err := m.pool.Get(time.Duration(DefaultPoolTimeout * time.Second))
	if err != nil {
		return r, fmt.Errorf("could not fetch module from pool - %s", err)
	}

	defer func() {
		err := m.pool.Return(i)
		if err != nil {
			defer i.Close(m.ctx)
		}
	}()

	r, err = i.Invoke(m.ctx, handler, payload)
	if err != nil {
		return r, fmt.Errorf("invocation of WASM module failed - %s", err)
	}

	return r, nil
}
