package metrics

import (
	"testing"

	proto "github.com/tarmac-project/protobuf-go/sdk/metrics"
	pb "google.golang.org/protobuf/proto"
)

type MetricsCase struct {
	Name     string
	Pass     bool
	Exists   bool
	Callback string
	Key      string
	Action   string
	Value    float64
}

func TestMetricsCallback(t *testing.T) {
	tc := []MetricsCase{
		{
			Name:     "Empty Metric Name Counter",
			Pass:     false,
			Exists:   false,
			Callback: "counter",
		},
		{
			Name:     "Empty Metric Name Gauge",
			Pass:     false,
			Exists:   false,
			Callback: "gauge",
		},
		{
			Name:     "Empty Metric Name Histogram",
			Pass:     false,
			Exists:   false,
			Callback: "histogram",
		},
		{
			Name:     "Happy Path",
			Pass:     true,
			Exists:   true,
			Callback: "counter",
			Key:      "happy_path",
		},
		{
			Name:     "Happy Path - Inc",
			Pass:     true,
			Exists:   true,
			Callback: "gauge",
			Key:      "happy_path_gauge",
			Action:   "inc",
		},
		{
			Name:     "Happy Path - Dec",
			Pass:     true,
			Exists:   true,
			Callback: "gauge",
			Key:      "happy_path_gauge",
			Action:   "dec",
		},
		{
			Name:     "Happy Path - Histogram",
			Pass:     true,
			Exists:   true,
			Callback: "histogram",
			Key:      "happy_path_histo",
			Value:    0.11231,
		},
		{
			Name:     "Weird Characters",
			Pass:     false,
			Exists:   false,
			Callback: "counter",
			Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
		},
		{
			Name:     "Weird Characters - Inc",
			Pass:     false,
			Exists:   false,
			Callback: "gauge",
			Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
			Action:   "inc",
		},
		{
			Name:     "Weird Characters - Dec",
			Pass:     false,
			Exists:   false,
			Callback: "gauge",
			Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
			Action:   "dec",
		},
		{
			Name:     "Weird Characters - Histogram",
			Pass:     false,
			Exists:   false,
			Callback: "histogram",
			Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
			Value:    0.11231,
		},
		{
			Name:     "Invalid Action",
			Pass:     false,
			Exists:   false,
			Callback: "gauge",
			Key:      "happy_path_gauge",
			Action:   "notvalid",
		},
		{
			Name:     "Missing Action",
			Pass:     false,
			Exists:   false,
			Callback: "gauge",
			Key:      "happy_path_gauge",
		},
		{
			Name:     "Missing Value",
			Pass:     false,
			Exists:   false,
			Callback: "histogram",
			Key:      "happy_path_histo",
		},
		{
			Name:     "Additional Runs",
			Pass:     true,
			Exists:   true,
			Callback: "counter",
			Key:      "happy_path",
		},
		{
			Name:     "Additional Runs - Inc",
			Pass:     true,
			Exists:   true,
			Callback: "gauge",
			Key:      "happy_path_gauge",
			Action:   "inc",
		},
		{
			Name:     "Additional Runs - Dec",
			Pass:     true,
			Exists:   true,
			Callback: "gauge",
			Key:      "happy_path_gauge",
			Action:   "dec",
		},
		{
			Name:     "Additional Runs - Histogram",
			Pass:     true,
			Exists:   true,
			Callback: "histogram",
			Key:      "happy_path_histo",
			Value:    0.11231,
		},
		{
			Name:     "Duplicate Name with different type - Counter",
			Pass:     false,
			Exists:   false,
			Callback: "counter",
			Key:      "happy_path_histo",
		},
		{
			Name:     "Duplicate Name with different type - Gauge",
			Pass:     false,
			Exists:   false,
			Callback: "gauge",
			Key:      "happy_path",
			Action:   "inc",
		},
		{
			Name:     "Duplicate Name with different type - Histogram",
			Pass:     false,
			Exists:   false,
			Callback: "histogram",
			Key:      "happy_path_gauge",
			Value:    0.11231,
		},
	}

	statsCallback, err := New(Config{})
	if err != nil {
		t.Fatalf("Unable to initialize new metrics - %s", err)
	}

	for _, c := range tc {
		t.Run(c.Name+" - "+c.Callback, func(t *testing.T) {
			switch c.Callback {
			case "counter":
				msg := &proto.MetricsCounter{
					Name: c.Key,
				}
				b, err := pb.Marshal(msg)
				if err != nil {
					t.Fatalf("Unable to marshal message - %s", err)
				}

				// Call Counter
				_, err = statsCallback.Counter(b)
				if err != nil && c.Pass {
					t.Errorf("Unexpected error calling callback - %s", err)
				}

				// Verify metric exists
				_, ok := statsCallback.counters[c.Key]
				if c.Exists && !ok {
					t.Errorf("Metric not created")
				}
			case "gauge":
				msg := &proto.MetricsGauge{
					Name:   c.Key,
					Action: c.Action,
				}
				b, err := pb.Marshal(msg)
				if err != nil {
					t.Fatalf("Unable to marshal message - %s", err)
				}

				// Call Gauge
				_, err = statsCallback.Gauge(b)
				if err != nil && c.Pass {
					t.Errorf("Unexpected error calling callback - %s", err)
				}

				// Verify metric exists
				_, ok := statsCallback.gauges[c.Key]
				if c.Exists && !ok {
					t.Errorf("Metric not created")
				}
			case "histogram":
				msg := &proto.MetricsHistogram{
					Name:  c.Key,
					Value: c.Value,
				}
				b, err := pb.Marshal(msg)
				if err != nil {
					t.Fatalf("Unable to marshal message - %s", err)
				}

				// Call Histogram
				_, err = statsCallback.Histogram(b)
				if err != nil && c.Pass {
					t.Errorf("Unexpected error calling callback - %s", err)
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

type MetricsCaseJSON struct {
	Name     string
	Pass     bool
	Exists   bool
	Callback string
	Key      string
	JSON     string
}

func TestMetricsCallbackJSON(t *testing.T) {
	var tc []MetricsCaseJSON

	tc = append(tc, MetricsCaseJSON{
		Name:     "Empty Metric Name",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "",
		JSON:     `{"name":""}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Happy Path",
		Pass:     true,
		Callback: "counter",
		Exists:   true,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Weird Characters",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
		JSON:     `{"name":"aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Invalid JSON",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Empty Metric Name",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "",
		JSON:     `{"name":""}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Invalid JSON",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Missing Action",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Duplicate Name",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path","action":"inc"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Happy Path - Inc",
		Pass:     true,
		Callback: "gauge",
		Exists:   true,
		Key:      "json_happy_path_gauge",
		JSON:     `{"name":"json_happy_path_gauge","action":"inc"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Happy Path - Dec",
		Pass:     true,
		Callback: "gauge",
		Exists:   true,
		Key:      "json_happy_path_gauge",
		JSON:     `{"name":"json_happy_path_gauge","action":"dec"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Invalid Action",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "json_happy_path_gauge",
		JSON:     `{"name":"json_happy_path_gauge","action":"notvalid"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Weird Characters",
		Pass:     false,
		Callback: "gauge",
		Exists:   false,
		Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
		JSON:     `{"name":"aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Empty Metric Name",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "",
		JSON:     `{"name":""}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Invalid JSON",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Missing Value",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "json_happy_path_histo",
		JSON:     `{"name":"json_happy_path"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Duplicate Name",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path","value":0.11231}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Happy Path",
		Pass:     true,
		Callback: "histogram",
		Exists:   true,
		Key:      "json_happy_path_histo",
		JSON:     `{"name":"json_happy_path_histo","value":0.11231}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Weird Characters",
		Pass:     false,
		Callback: "histogram",
		Exists:   false,
		Key:      "aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ",
		JSON:     `{"name":"aw3er2324re2309vcqASEDFAQSWWEqwrqwQ!@#$@#VQ"}`,
	})

	tc = append(tc, MetricsCaseJSON{
		Name:     "Duplicate Name",
		Pass:     false,
		Callback: "counter",
		Exists:   false,
		Key:      "json_happy_path",
		JSON:     `{"name":"json_happy_path_histo"}`,
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
