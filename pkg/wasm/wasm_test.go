package wasm

import (
	"context"
	"testing"
	"time"
)

func TestWASMServerCreation(t *testing.T) {
	_, err := NewServer(Config{})
	if err == nil {
		t.Errorf("New Server creation should have errored with no Callback")
	}
}

type ModuleCase struct {
	ModuleConf ModuleConfig
	Pass       bool
	Name       string
}

func TestWASMModuleCreation(t *testing.T) {
	s, err := NewServer(Config{
		Callback: func(context.Context, string, string, string, []byte) ([]byte, error) { return []byte(""), nil },
	})
	if err != nil {
		t.Errorf("Failed to create WASM Server - %s", err)
	}
	defer s.Shutdown()

	var mc []ModuleCase

	// Happy Path
	mc = append(mc, ModuleCase{
		Name: "Happy Path",
		Pass: true,
		ModuleConf: ModuleConfig{
			Name:     "A Module",
			PoolSize: 99,
			Filepath: "/testdata/logger/tarmac.wasm",
		},
	})

	// No Name
	mc = append(mc, ModuleCase{
		Name: "No Name",
		Pass: false,
		ModuleConf: ModuleConfig{
			PoolSize: 99,
			Filepath: "/testdata/logger/tarmac.wasm",
		},
	})

	// No Pool Size
	mc = append(mc, ModuleCase{
		Name: "No Pool Size",
		Pass: true,
		ModuleConf: ModuleConfig{
			Name:     "A Module",
			Filepath: "/testdata/logger/tarmac.wasm",
		},
	})

	// No File
	mc = append(mc, ModuleCase{
		Name: "No File",
		Pass: false,
		ModuleConf: ModuleConfig{
			Name:     "A Module",
			PoolSize: 99,
		},
	})

	// Bad File Location
	mc = append(mc, ModuleCase{
		Name: "Bad File Location",
		Pass: false,
		ModuleConf: ModuleConfig{
			Name:     "A Module",
			PoolSize: 99,
			Filepath: "/doesntexist/testdata/logger/tarmac.wasm",
		},
	})

	// Execute Test Cases
	for _, m := range mc {
		t.Run("Module Creation Test Case - "+m.Name, func(t *testing.T) {
			err := s.LoadModule(m.ModuleConf)
			if !m.Pass && err == nil {
				t.Errorf("Case should have failed, but it passed")
			}
			if m.Pass && err != nil {
				t.Errorf("Case should have passed, but it failed - %s", err)
			}
		})
	}

	// Try to lookup a non-existent module
	t.Run("Non-existent module lookup", func(t *testing.T) {
		_, err := s.Module("ThisBetterFail")
		if err == nil {
			t.Fatalf("Non-existent module lookup succeeded...")
		}
	})
}

func TestWASMExecution(t *testing.T) {
	callbackCh := make(chan struct{}, 2)
	s, err := NewServer(Config{
		Callback: func(context.Context, string, string, string, []byte) ([]byte, error) {
			callbackCh <- struct{}{}
			return []byte(""), nil
		},
	})
	if err != nil {
		t.Fatalf("Failed to create WASM Server - %s", err)
	}
	defer s.Shutdown()

	err = s.LoadModule(ModuleConfig{
		Name:     "AModule",
		Filepath: "/testdata/logger/tarmac.wasm",
	})
	if err != nil {
		t.Fatalf("Failed to load module - %s", err)
	}

	m, err := s.Module("AModule")
	if err != nil {
		t.Fatalf("Cannot find module - %s - %+v", err, s)
	}

	go func() {
		_, err = m.Run("handler", []byte(`hello`))
		if err != nil {
			t.Logf("Could not execute the wasm function - %s", err)
		}
	}()

	// Check for Callback execution
	select {
	case <-time.After(15 * time.Second):
		t.Errorf("Timeout waiting for Callback execution")
	case <-callbackCh:
		return
	}
}
