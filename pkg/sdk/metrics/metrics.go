/*
Package metrics is a client package for WASM functions running within a Tarmac server.

This package provides user-friendly Metrics functions that enable users to create custom metrics. Guest WASM functions running inside Tarmac can import and call this Metrics interface.
*/
package metrics

import "fmt"

// Metrics provides an interface to the host metrics.
type Metrics struct {
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// Config provides users with the ability to specify namespaces, function handlers and other key information required to execute the
// function.
type Config struct {
	// Namespace controls the function namespace to use for host callbacks. The default value is "default" which is the global namespace.
	// Users can provide an alternative namespace by specifying this field.
	Namespace string

	// HostCall is used internally for host callbacks. This is mainly here for testing.
	HostCall func(string, string, string, []byte) ([]byte, error)
}

// New creates a metrics interface for creating Counters, Histograms, and Gauges.
func New(cfg Config) (*Metrics, error) {
	// Set default namespace
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// Verify HostCall is set
	if cfg.HostCall == nil {
		return &Metrics{}, fmt.Errorf("HostCall cannot be nil")
	}

	return &Metrics{namespace: cfg.Namespace, hostCall: cfg.HostCall}, nil
}

// Counter provides an Counter metric which can be incremented.
type Counter struct {
	name      string
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// NewCounter creates a new Counter metric with the given name.
func (m *Metrics) NewCounter(name string) (*Counter, error) {
	if name == "" {
		return &Counter{}, fmt.Errorf("name cannot be empty")
	}

	c := &Counter{name: name, namespace: m.namespace, hostCall: m.hostCall}
	return c, nil
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	_, _ = c.hostCall(c.namespace, "metrics", "counter", []byte(fmt.Sprintf(`{"name":"%s"}`, c.name)))
}

// Histogram provides a Histogram metric which can observe values.
type Histogram struct {
	name      string
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// NewHistogram creates a new Histogram metric with the given name.
func (m *Metrics) NewHistogram(name string) (*Histogram, error) {
	if name == "" {
		return &Histogram{}, fmt.Errorf("name cannot be empty")
	}

	h := &Histogram{name: name, namespace: m.namespace, hostCall: m.hostCall}
	return h, nil
}

// Observe adds a value to the histogram.
func (h *Histogram) Observe(f float64) {
	_, _ = h.hostCall(h.namespace, "metrics", "histogram", []byte(fmt.Sprintf(`{"name":"%s","value": %f}`, h.name, f)))
}

// Gauge provides a Gauge metric which can be incremented and decremented.
type Gauge struct {
	name      string
	namespace string
	hostCall  func(string, string, string, []byte) ([]byte, error)
}

// NewGauge creates a new Gauge metric with the given name.
func (m *Metrics) NewGauge(name string) (*Gauge, error) {
	if name == "" {
		return &Gauge{}, fmt.Errorf("name cannot be empty")
	}

	g := &Gauge{name: name, namespace: m.namespace, hostCall: m.hostCall}
	return g, nil
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	_, _ = g.hostCall(g.namespace, "metrics", "gauge", []byte(fmt.Sprintf(`{"name":"%s","action": "inc"}`, g.name)))
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	_, _ = g.hostCall(g.namespace, "metrics", "gauge", []byte(fmt.Sprintf(`{"name":"%s","action": "dec"}`, g.name)))
}
