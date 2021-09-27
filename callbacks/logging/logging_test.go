package logging

import (
	"testing"
)

func TestLoggingDefaults(t *testing.T) {
	l, err := New(Config{})
	if err != nil {
		t.Fatalf("Unable to create new logger - %s", err)
	}

	t.Run("Info Logging", func(t *testing.T) {
		b, err := l.Info([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Error Logging", func(t *testing.T) {
		b, err := l.Error([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Warn Logging", func(t *testing.T) {
		b, err := l.Warn([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Debug Logging", func(t *testing.T) {
		b, err := l.Debug([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Trace Logging", func(t *testing.T) {
		b, err := l.Trace([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})
}

func TestLoggingFunc(t *testing.T) {
	log := NoOpLog{}
	l, err := New(Config{Log: log})
	if err != nil {
		t.Fatalf("Unable to create new logger - %s", err)
	}

	t.Run("Info Logging", func(t *testing.T) {
		b, err := l.Info([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Error Logging", func(t *testing.T) {
		b, err := l.Error([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Warn Logging", func(t *testing.T) {
		b, err := l.Warn([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Debug Logging", func(t *testing.T) {
		b, err := l.Debug([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Trace Logging", func(t *testing.T) {
		b, err := l.Trace([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

}
