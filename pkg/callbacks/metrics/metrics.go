/*
Package metrics is part of the Tarmac suite of Host Callback packages. This package provides users with the ability
to provide WASM functions with a host callback interface that provides metrics tracking capabilities.

	import (
		"github.com/tarmac-project/tarmac/pkg/callbacks/metrics"
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
	"regexp"
	"sync"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tarmac-project/tarmac"

	proto "github.com/tarmac-project/protobuf-go/sdk/metrics"
	pb "google.golang.org/protobuf/proto"
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
func New(_ Config) (*Metrics, error) {
	m := &Metrics{}
	m.all = make(map[string]string)
	m.counters = make(map[string]prometheus.Counter)
	m.gauges = make(map[string]prometheus.Gauge)
	m.histograms = make(map[string]prometheus.Summary)
	return m, nil
}

func (m *Metrics) Counter(b []byte) ([]byte, error) {
	rq := &proto.MetricsCounter{}
	err := pb.Unmarshal(b, rq)
	if err != nil {
		return m.jsonCounter(b)
	}

	return []byte(""), m.counter(rq.Name)
}

func (m *Metrics) jsonCounter(b []byte) ([]byte, error) {
	// Parse incoming Request
	var rq tarmac.MetricsCounter
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to parse input JSON - %s", err)
	}

	return []byte(""), m.counter(rq.Name)
}

func (m *Metrics) counter(name string) error {
	// Verify Name Value
	if !isMetricNameValid.MatchString(name) {
		return ErrInvalidMetricName
	}

	// Map Safety
	m.Lock()
	defer m.Unlock()

	// Check if counter already exists, if not create one
	_, ok := m.counters[name]
	if !ok {
		// Check if name is already used
		_, ok2 := m.all[name]
		if ok2 {
			return fmt.Errorf("metric name in use")
		}
		m.counters[name] = promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
		})
		m.all[name] = "counter"
	}

	// Perform action
	m.counters[name].Inc()
	return nil
}

// Gauge will create a gauge metric and perform the provided action.
func (m *Metrics) Gauge(b []byte) ([]byte, error) {
	// Parse incoming Request
	rq := &proto.MetricsGauge{}
	err := pb.Unmarshal(b, rq)
	if err != nil {
		return m.jsonGauge(b)
	}

	return []byte(""), m.gauge(rq.Name, rq.Action)
}

func (m *Metrics) jsonGauge(b []byte) ([]byte, error) {
	// Parse incoming Request
	var rq tarmac.MetricsGauge
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to parse input JSON - %s", err)
	}

	return []byte(""), m.gauge(rq.Name, rq.Action)
}

func (m *Metrics) gauge(name string, action string) error {
	// Verify Name Value
	if !isMetricNameValid.MatchString(name) {
		return ErrInvalidMetricName
	}

	if action != "inc" && action != "dec" {
		return fmt.Errorf("invalid action")
	}

	// Map Safety
	m.Lock()
	defer m.Unlock()

	// Check if gauge already exists, if not create one
	_, ok := m.gauges[name]
	if !ok {
		// Check if name is already used
		_, ok2 := m.all[name]
		if ok2 {
			return fmt.Errorf("metric name in use")
		}
		m.gauges[name] = promauto.NewGauge(prometheus.GaugeOpts{
			Name: name,
		})
		m.all[name] = "gauge"
	}

	// Perform action
	switch action {
	case "inc":
		m.gauges[name].Inc()
	case "dec":
		m.gauges[name].Dec()
	default:
		return fmt.Errorf("invalid action")
	}

	return nil
}

// Histogram will create a histogram metric and perform the provided action.
func (m *Metrics) Histogram(b []byte) ([]byte, error) {
	// Parse incoming Request
	rq := &proto.MetricsHistogram{}
	err := pb.Unmarshal(b, rq)
	if err != nil {
		return m.jsonHistogram(b)
	}

	return []byte(""), m.histogram(rq.Name, rq.Value)
}

func (m *Metrics) jsonHistogram(b []byte) ([]byte, error) {
	// Parse incoming Request
	var rq tarmac.MetricsHistogram
	err := ffjson.Unmarshal(b, &rq)
	if err != nil {
		return []byte(""), fmt.Errorf("unable to parse input JSON - %s", err)
	}

	return []byte(""), m.histogram(rq.Name, rq.Value)
}

func (m *Metrics) histogram(name string, value float64) error {
	// Verify Name Value
	if !isMetricNameValid.MatchString(name) {
		return ErrInvalidMetricName
	}

	// Map Safety
	m.Lock()
	defer m.Unlock()

	// Check if histogram already exists, if not create one
	_, ok := m.histograms[name]
	if !ok {
		// Check if name is already used
		_, ok2 := m.all[name]
		if ok2 {
			return fmt.Errorf("metric name in use")
		}
		m.histograms[name] = promauto.NewSummary(prometheus.SummaryOpts{
			Name:       name,
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})
		m.all[name] = "histogram"
	}

	// Perform action
	m.histograms[name].Observe(value)
	return nil
}
