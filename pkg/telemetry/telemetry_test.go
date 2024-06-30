package telemetry

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewTelemetry(t *testing.T) {
	// Test Multiple Initializations and Closures do not panic
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("TestNewTelemetry - %d", i), func(_ *testing.T) {
			<-time.After(1 * time.Second)
			// Create a new Telemetry instance
			tm := New()
			defer tm.Close()

			// Simulate usage of the metrics
			labels := prometheus.Labels{"path": "/api"}
			tm.Srv.With(labels).Observe(0.5)

			taskLabels := prometheus.Labels{"tasks": "task1"}
			tm.Tasks.With(taskLabels).Observe(1.2)

			callbackLabels := prometheus.Labels{"callback": "callback1"}
			tm.Callbacks.With(callbackLabels).Observe(0.8)

			wasmLabels := prometheus.Labels{"route": "/wasm"}
			tm.Wasm.With(wasmLabels).Observe(0.3)

			routeLabels := prometheus.Labels{"service": "service1", "type": "http"}
			tm.Routes.With(routeLabels).Inc()
			tm.Routes.With(routeLabels).Dec()

		})
	}
}
