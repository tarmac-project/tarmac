package telemetry

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTelemetry(t *testing.T) {
	// Test Multiple Initializations and Closures do not panic
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("TestNewTelemetry - %d", i), func(t *testing.T) {
			<-time.After(1 * time.Second)
			// Create a new Telemetry instance
			tm := New()
			defer tm.Close()

			// Check if the Telemetry instance is not nil
			if tm == nil {
				t.Error("New Telemetry instance is nil")
			}

			// Check if the Tasks field is not nil
			if tm.Tasks == nil {
				t.Error("Tasks field is nil")
			}

			// Check if the Srv field is not nil
			if tm.Srv == nil {
				t.Error("Srv field is nil")
			}

			// Check if the Callbacks field is not nil
			if tm.Callbacks == nil {
				t.Error("Callbacks field is nil")
			}

			// Check if the Wasm field is not nil
			if tm.Wasm == nil {
				t.Error("Wasm field is nil")
			}
		})
	}
}
