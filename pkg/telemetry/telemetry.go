/*
Package telemetry provides the capability to manage and hold system internal metrics.
*/
package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Telemetry provides the capability to manage and hold system internal metrics.
type Telemetry struct {
	// Tasks is a summary metric of user scheduled task executions.
	Tasks *prometheus.SummaryVec

	// Srv is a summary metric of the HTTP server request processing.
	Srv *prometheus.SummaryVec

	// Callbacks is a summary metric of the WASM callbacks guests make.
	Callbacks *prometheus.SummaryVec

	// Wasm is a summary metric of the WASM guest module executions.
	Wasm *prometheus.SummaryVec

	// Routes is a gauge metric of the configured service routes.
	Routes *prometheus.GaugeVec
}

// New creates and returns an initialized Telemetry instance with default metrics.
func New() *Telemetry {
	m := &Telemetry{}

	m.Srv = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "http_server",
		Help:       "Summary of HTTP Server requests",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"path"},
	)

	m.Tasks = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "scheduled_tasks",
		Help:       "Summary of user defined scheduled task WASM function executions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"tasks"},
	)

	m.Callbacks = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "wasm_callbacks",
		Help:       "Summary of server callbacks from WASM function executions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"callback"},
	)

	m.Wasm = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "wasm_functions",
		Help:       "Summary of wasm function executions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"route"},
	)

	m.Routes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "service_routes",
		Help: "Number of configured service routes",
	},
		[]string{"service", "type"},
	)

	return m
}

// Close unregisters the Telemetry metrics from the Prometheus registry.
func (t *Telemetry) Close() {
	_ = prometheus.Unregister(t.Tasks)
	_ = prometheus.Unregister(t.Srv)
	_ = prometheus.Unregister(t.Callbacks)
	_ = prometheus.Unregister(t.Wasm)
	_ = prometheus.Unregister(t.Routes)
}
