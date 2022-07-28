/*
Telemetry is an internal Tarmac package used to initialize system metrics.
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
}

// New will return an initialized systems metrics instance.
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

	return m
}
