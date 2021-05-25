package app

import (
    "fmt"
    wasmer "github.com/wasmerio/wasmer-go/wasmer"
    "io/ioutil"
    "sync"
)

// WASMServer is used as a WASM runtime engine providing capabilities to load and run modules.
type WASMServer struct {
  sync.RWMutex
  Engine *wasmer.Engine
  Store *wasmer.Store
  modules map[string]*WASMModule
}

// WASMModule is a specific WASM Module that can be loaded into the engine.
type WASMModule struct {
  Name  string
  Module  *wasmer.Module
}

// NewWASMServer will create a new WASMServer with the Engine and Module Store pre-loaded.
func NewWASMServer() (*WASMServer, error) {
  s := &WASMServer{}
  s.Engine = wasmer.NewEngine()
  s.Store = wasmer.NewStore(s.Engine)
  s.modules = make(map[string]*WASMModule)
  return s, nil
}

// LoadModule will read and load the WASM module from the filesysem.
func (s *WASMServer) LoadModule(key, file string) error {
  if key == "" || file == "" {
    return fmt.Errorf("key and file cannot be empty")
  }

  // Read the WASM module file
  bytes, err := ioutil.ReadFile(file)
  if err != nil {
    return fmt.Errorf("unable to read wasm module file - %s", err)
  }

  // Create a new wasmer Module with file
  module, err := wasmer.NewModule(s.Store, bytes)
  if err != nil {
    return fmt.Errorf("unable to create new wasmer module with wasm file %s - %s", file, err)
  }

  // Add Module into local map
  s.Lock()
  defer s.Unlock()
  s.modules[key] = &WASMModule{
    Name: key,
    Module: module,
  }

  return nil
}

// Module will return the WASMModule stored for the specified WASM module.
func (s *WASMServer) Module(key string) *WASMModule {
  s.RLock()
  defer s.RUnlock()
  return s.modules[key]
}
