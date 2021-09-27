/*
Package metrics is part of the Tarmac suite of Host Callback packages. This package provides users with the ability
to provide WASM functions with a host callback interface that provides metrics tracking capabilities.

	package main

	import (
		"github.com/madflojo/tarmac/callbacks"
		"github.com/madflojo/tarmac/callbacks/metrics"
	)

	func main() {
		// Create instance of metrics to register for callback execution
		metrics := metrics.New(metrics.Config{})

		// Create Callback router and register metrics
		router := callbacks.New()
		router.RegisterCallback("metrics", "Counter", metrics.Counter)
	}


*/
package metrics

import (
	"fmt"
	"github.com/madflojo/tarmac"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"regexp"
	"sync"
)

// Metrics stores and manages the user-defined metrics created via
// WASM function callbacks.
type Metrics struct {
	sync.Mutex
	// all contains a map of all custom defined metrics
	all map[string]string

	// counters holds a map of existing custom counters
	counters map[string]prometheus.Counter

	// gauges holds a map of existing custom gauges
	gauges map[string]prometheus.Gauge

	// histograms holds a map of existing custom summaries
	histograms map[string]prometheus.Summary
}

// ErrInvalidMetricName is an error returned when the user supplies an
// invalid formatted metric name.
var ErrInvalidMetricName = fmt.Errorf("invalid metric name")

// isMetricNameValid is a regex used to validate metric names.
var isMetricNameValid = regexp.MustCompile(`^[a-zA-Z0-9_:][a-zA-Z0-9_:]*$`)

// Config is provided to users to configure the Host Callback. All Tarmac Callbacks follow the same configuration
// format; each Config struct gives the specific Host Callback unique functionality.
type Config struct{}

// New will create a new instance of metrics enabling users to
// collect custom metrics.
func New(cfg Config) (*Metrics, error) {
	m := &Metrics{}
	m.all = make(map[string]string)
	m.counters = make(map[string]prometheus.Counter)
	m.gauges = make(map[string]prometheus.Gauge)
	m.histograms = make(map[string]prometheus.Summary)
	return m, nil
}

// Counter will create and increment a counter metric. The expected input for
// this function is a MetricsCounter JSON.
func (m *Metrics) Counter(b []byte) ([]byte, error) {
	// Parse incoming Request
	var rq tarmac.MetricsCounter
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to parse input JSON - %s", err)
	}

	// Verify Name Value
	if !isMetricNameValid.MatchString(rq.Name) {
		return []byte(""), ErrInvalidMetricName
	}

	// Map Safety
	m.Lock()
	defer m.Unlock()

	// Check if counter already exists, if not create one
	_, ok := m.counters[rq.Name]
	if !ok {
		// Check if name is already used
		_, ok2 := m.all[rq.Name]
		if ok2 {
			return []byte(""), fmt.Errorf("metric name in use")
		}
		m.counters[rq.Name] = promauto.NewCounter(prometheus.CounterOpts{
			Name: rq.Name,
		})
		m.all[rq.Name] = "counter"
	}

	// Perform action
	m.counters[rq.Name].Inc()
	return []byte(""), nil
}

// Guage will create a gauge metric and either increment or decrement the value
// based on the provided input. The expected input for this function is a
// MetricsGauge JSON.
func (m *Metrics) Gauge(b []byte) ([]byte, error) {
	// Parse incoming Request
	var rq tarmac.MetricsGauge
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to parse input JSON - %s", err)
	}

	// Verify Name Value
	if !isMetricNameValid.MatchString(rq.Name) {
		return []byte(""), ErrInvalidMetricName
	}

	// Map Safety
	m.Lock()
	defer m.Unlock()

	// Check if gauge already exists, if not create one
	_, ok := m.gauges[rq.Name]
	if !ok {
		// Check if name is already used
		_, ok2 := m.all[rq.Name]
		if ok2 {
			return []byte(""), fmt.Errorf("metric name in use")
		}
		m.gauges[rq.Name] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: rq.Name,
		})
		m.all[rq.Name] = "gauge"
	}

	// Perform action
	switch rq.Action {
	case "inc":
		m.gauges[rq.Name].Inc()
	case "dec":
		m.gauges[rq.Name].Dec()
	default:
		return []byte(""), fmt.Errorf("invalid action")
	}

	return []byte(""), nil
}

// Histogram will create a histogram or summary metric and observe the
// provided values. The expected input for this function is a
// MetricsHistogram JSON.
func (m *Metrics) Histogram(b []byte) ([]byte, error) {
	// Parse incoming Request
	var rq tarmac.MetricsHistogram
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to parse input JSON - %s", err)
	}

	// Verify Name Value
	if !isMetricNameValid.MatchString(rq.Name) {
		return []byte(""), ErrInvalidMetricName
	}

	// Map Safety
	m.Lock()
	defer m.Unlock()

	// Check if histogram already exists, if not create one
	_, ok := m.histograms[rq.Name]
	if !ok {
		// Check if name is already used
		_, ok2 := m.all[rq.Name]
		if ok2 {
			return []byte(""), fmt.Errorf("metric name in use")
		}
		m.histograms[rq.Name] = promauto.NewSummary(prometheus.SummaryOpts{
			Name:       rq.Name,
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})
		m.all[rq.Name] = "histogram"
	}

	// Perform action
	m.histograms[rq.Name].Observe(rq.Value)

	return []byte(""), nil
}
