package telemetry

import (
	"testing"
)

func NewTelemetry(_ *testing.T) {
	m := New()
	m.Srv.Reset()
}
