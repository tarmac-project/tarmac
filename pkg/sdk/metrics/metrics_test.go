package metrics

import (
	"testing"

	"github.com/pquerna/ffjson/ffjson"
)

type CounterTestCase struct {
	caseName string
	name     string
	nameErr  bool
	hostCall func(string, string, string, []byte) ([]byte, error)
}

func TestMetricsCounters(t *testing.T) {
	tt := []CounterTestCase{
		{
			caseName: "Valid Metric Name",
			name:     "testing",
			nameErr:  false,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				if namespace != "default" || capability != "metrics" || function != "counter" {
					t.Errorf("invalid hostCall signature - %s %s %s", namespace, capability, function)
				}
				return []byte(""), nil
			},
		},
		{
			caseName: "Empty Metric Name",
			name:     "",
			nameErr:  true,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				t.Errorf("hostCall should not have been called")
				return []byte(""), nil
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {
			// Create new metrics instance using table inputs
			m, err := New(Config{HostCall: tc.hostCall, Namespace: "default"})
			if err != nil {
				t.Errorf("unexpected error creating metric instance - %s", err)
			}

			// Create new counter using table inputs
			counter, err := m.NewCounter(tc.name)
			if err != nil && tc.nameErr {
				return
			}
			if err != nil && !tc.nameErr {
				t.Errorf("counter creation returned unexpected error - %s", err)
				return
			}
			if err == nil && tc.nameErr {
				t.Errorf("counter creation did not return expected error")
				return
			}

			// Call increment
			counter.Inc()
		})
	}
}

type GaugeTestCase struct {
	caseName string
	name     string
	nameErr  bool
	hostCall func(string, string, string, []byte) ([]byte, error)
}

func TestMetricsGauges(t *testing.T) {
	tt := []GaugeTestCase{
		{
			caseName: "Valid Metric Name",
			name:     "testing",
			nameErr:  false,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				if namespace != "default" || capability != "metrics" || function != "gauge" {
					t.Errorf("invalid hostCall signature - %s %s %s", namespace, capability, function)
				}
				return []byte(""), nil
			},
		},
		{
			caseName: "Empty Metric Name",
			name:     "",
			nameErr:  true,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				t.Errorf("hostCall should not have been called")
				return []byte(""), nil
			},
		},
		{
			caseName: "Increment",
			name:     "testing",
			nameErr:  false,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				if namespace != "default" || capability != "metrics" || function != "gauge" {
					t.Errorf("invalid hostCall signature - %s %s %s", namespace, capability, function)
				}
				if string(input) != `{"name":"testing","action": "inc"}` {
					t.Errorf("invalid input - %s", input)
				}
				return []byte(""), nil
			},
		},
		{
			caseName: "Decrement",
			name:     "testing",
			nameErr:  false,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				if namespace != "default" || capability != "metrics" || function != "gauge" {
					t.Errorf("invalid hostCall signature - %s %s %s", namespace, capability, function)
				}
				if string(input) != `{"name":"testing","action": "dec"}` {
					t.Errorf("invalid input - %s", input)
				}
				return []byte(""), nil
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {
			// Create new metrics instance using table inputs
			m, err := New(Config{HostCall: tc.hostCall, Namespace: "default"})
			if err != nil {
				t.Errorf("unexpected error creating metric instance - %s", err)
			}

			// Create new gauge using table inputs
			gauge, err := m.NewGauge(tc.name)
			if err != nil && tc.nameErr {
				return
			}
			if err != nil && !tc.nameErr {
				t.Errorf("gauge creation returned unexpected error - %s", err)
				return
			}
			if err == nil && tc.nameErr {
				t.Errorf("gauge creation did not return expected error")
				return
			}

			// Call increment and decrement
			// Call Inc or Dec
			if tc.caseName == "Increment" {
				gauge.Inc()
			} else if tc.caseName == "Decrement" {
				gauge.Dec()
			}
		})
	}
}

type HistogramTestCase struct {
	caseName string
	name     string
	nameErr  bool
	hostCall func(string, string, string, []byte) ([]byte, error)
}

func TestMetricsHistogram(t *testing.T) {
	tt := []HistogramTestCase{
		{
			caseName: "Valid Metric Name and Value",
			name:     "testing",
			nameErr:  false,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				if namespace != "default" || capability != "metrics" || function != "histogram" {
					t.Errorf("invalid hostCall signature - %s %s %s", namespace, capability, function)
				}

				var req map[string]interface{}
				err := ffjson.Unmarshal(input, &req)
				if err != nil {
					t.Errorf("unexpected error parsing input JSON: %v", err)
				}

				if req["name"] != "testing" {
					t.Errorf("expected input to contain name: %v", req)
				}
				if req["value"] != 1.23 {
					t.Errorf("expected input to contain value: %v", req)
				}

				return []byte(""), nil
			},
		},
		{
			caseName: "Empty Metric Name",
			name:     "",
			nameErr:  true,
			hostCall: func(namespace, capability, function string, input []byte) ([]byte, error) {
				t.Errorf("hostCall should not have been called")
				return []byte(""), nil
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.caseName, func(t *testing.T) {
			// Create new metrics instance using table inputs
			m, err := New(Config{HostCall: tc.hostCall, Namespace: "default"})
			if err != nil {
				t.Errorf("unexpected error creating metric instance - %s", err)
			}

			// Create new histogram using table inputs
			histogram, err := m.NewHistogram(tc.name)
			if err != nil && tc.nameErr {
				return
			}
			if err != nil && !tc.nameErr {
				t.Errorf("histogram creation returned unexpected error - %s", err)
				return
			}
			if err == nil && tc.nameErr {
				t.Errorf("histogram creation did not return expected error")
				return
			}

			// Call observe with a value
			histogram.Observe(1.23)
		})
	}
}
