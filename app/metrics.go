package app

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// telemetry provides the capability to manage and hold system internal metrics.
type telemetry struct {
	// tasks is a summary metric of user scheduled task executions.
	tasks *prometheus.SummaryVec

	// srv is a summary metric of the HTTP server request processing.
	srv *prometheus.SummaryVec

	// callbacks is a summary metric of the WASM callbacks guests make.
	callbacks *prometheus.SummaryVec

	// wasm is a summary metric of the WASM guest module executions.
	wasm *prometheus.SummaryVec
}

// newTelemetry will return an initialized systems metrics instance.
func newTelemetry() *telemetry {
	m := &telemetry{}

	m.srv = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "http_server",
		Help:       "Summary of HTTP Server requests",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"path"},
	)

	m.tasks = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "scheduled_tasks",
		Help:       "Summary of user defined scheduled task WASM function executions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"tasks"},
	)

	m.callbacks = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "wasm_callbacks",
		Help:       "Summary of server callbacks from WASM function executions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"callback"},
	)

	m.wasm = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "wasm_functions",
		Help:       "Summary of wasm function executions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	},
		[]string{"route"},
	)

	return m
}
