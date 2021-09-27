package metrics

import (
	"testing"
)

type MetricsCase struct {
	Name     string
	Pass     bool
	Exists   bool
	Callback string
	Key      string
	JSON     string
}

func TestMetricsCallback(t *testing.T) {
	var tc []MetricsCase

	tc = append(tc, MetricsCase{
		Name:     "Empty Metric Name",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "",
		JSON:     `{"name":""}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Happy Path",
		Pass:     true,
		Callback: "counter",
		Exists:   true,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Weird Characters",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
		JSON:     `{"name":"aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Invalid JSON",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Empty Metric Name",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "",
		JSON:     `{"name":""}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Invalid JSON",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Missing Action",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Duplicate Name",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path","action":"inc"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Happy Path - Inc",
		Pass:     true,
		Callback: "gauge",
		Exists:   true,
		Key:      "happy_path_gauge",
		JSON:     `{"name":"happy_path_gauge","action":"inc"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Happy Path - Dec",
		Pass:     true,
		Callback: "gauge",
		Exists:   true,
		Key:      "happy_path_gauge",
		JSON:     `{"name":"happy_path_gauge","action":"dec"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Invalid Action",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "happy_path_gauge",
		JSON:     `{"name":"happy_path_gauge","action":"notvalid"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Weird Characters",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
		JSON:     `{"name":"aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Empty Metric Name",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "",
		JSON:     `{"name":""}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Invalid JSON",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Missing Value",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "happy_path_histo",
		JSON:     `{"name":"happy_path"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Duplicate Name",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path","Value":0.11231}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Happy Path",
		Pass:     true,
		Callback: "histogram",
		Exists:   true,
		Key:      "happy_path_histo",
		JSON:     `{"name":"happy_path_histo","Value":0.11231}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Weird Characters",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
		JSON:     `{"name":"aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ"}`,
	})

	tc = append(tc, MetricsCase{
		Name:     "Duplicate Name",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "happy_path",
		JSON:     `{"name":"happy_path_histo"}`,
	})

	statsCallback, err := New(Config{})
	if err != nil {
		t.Fatalf("Unable to initialize new metrics - %s", err)
	}

	for _, c := range tc {
		t.Run(c.Name+" - "+c.Callback, func(t *testing.T) {
			switch c.Callback {
			case "counter":
				// Call Counter
				_, err := statsCallback.Counter([]byte(c.JSON))
				if err != nil && c.Pass {
					t.Errorf("Unexpected error calling callback - %s", err)
				}
				if err == nil && !c.Pass {
					t.Errorf("Expected error calling callback - got nil")
				}

				// Verify metric exists
				_, ok := statsCallback.counters[c.Key]
				if c.Exists && !ok {
					t.Errorf("Metric not created")
				}
			case "gauge":
				// Call Gauge
				_, err := statsCallback.Gauge([]byte(c.JSON))
				if err != nil && c.Pass {
					t.Errorf("Unexpected error calling callback - %s", err)
				}
				if err == nil && !c.Pass {
					t.Errorf("Expected error calling callback - got nil")
				}

				// Verify metric exists
				_, ok := statsCallback.gauges[c.Key]
				if c.Exists && !ok {
					t.Errorf("Metric not created")
				}
			case "histogram":
				// Call Histogram
				_, err := statsCallback.Histogram([]byte(c.JSON))
				if err != nil && c.Pass {
					t.Errorf("Unexpected error calling callback - %s", err)
				}
				if err == nil && !c.Pass {
					t.Errorf("Expected error calling callback - got nil")
				}

				// Verify metric exists
				_, ok := statsCallback.histograms[c.Key]
				if c.Exists && !ok {
					t.Errorf("Metric not created")
				}
			default:
				t.Errorf("Unknown callback method - %s", c.Callback)
			}
		})
	}
}
