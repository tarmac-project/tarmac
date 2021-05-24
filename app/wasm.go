package app

import (
    "fmt"
    wasmer "github.com/wasmerio/wasmer-go/wasmer"
    "io/ioutil"
)

type WASMServer struct {
  Engine *wasmer.Engine
  Store *wasmer.Store
  modules map[string]*WASMModule
}

type WASMModule struct {
  Name  string
  Module  *wasmer.Module
}

func NewWASMServer() (*WASMServer, error) {
  s := &WASMServer{}
  s.Engine = wasmer.NewEngine()
  s.Store = wasmer.NewStore(s.Engine)
  s.modules = make(map[string]*WASMModule)
  return s, nil
}


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

  s.modules[key] = &WASMModule{
    Name: key,
    Module: module,
  }

  return nil
}

func (s *WASMServer) Module(key string) *WASMModule {
  return s.modules[key]
}
