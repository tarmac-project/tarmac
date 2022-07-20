package telemetry

import (
	"testing"
)

func NewTelemetry(t *testing.T) {
	m := New()
	m.Srv.Reset()
}
