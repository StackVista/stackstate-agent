package testrawmetrics

import (
	"fmt"
	"os"
	"testing"

	"github.com/StackVista/stackstate-agent/rtloader/test/helpers"
)

func TestMain(m *testing.M) {
	err := setUp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up tests: %v", err)
		os.Exit(-1)
	}

	os.Exit(m.Run())
}

func TestSubmitMetricsRawData(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	// submit_metrics_raw_data
	out, err := run(`rawmetrics.submit_raw_metrics_data(
							None,
							"checkid",
							[
								{
									"name": "ev_name",
									"timestamp": 15,
									"value": "ev_value",
									"hostname": "ev_hostname",
									"tags": [
										{
											"ev_key": "ev_value"
										}
									]
								}
							])
				`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}

	if len(_rawMetric) != 1 {
		t.Fatalf("Unexpected metric data, raw metrics array is required")
	}

	metric := _rawMetric[0]

	if metric.Name != "ev_name" {
		t.Fatalf("Unexpected metric data 'name' value: %s", metric.Name)
	}

	if metric.Timestamp != 15 {
		t.Fatalf("Unexpected metric data 'timestamp' value: %v", metric.Timestamp)
	}

	if metric.Value != "ev_value" {
		t.Fatalf("Unexpected metric data 'value' value: %s", metric.Value)
	}

	if metric.Hostname != "ev_hostname" {
		t.Fatalf("Unexpected metric data 'hostname' value: %s", metric.Hostname)
	}

	if len(metric.Tags) != 1 {
		t.Fatalf("Unexpected metric data 'tags' size: %v", len(metric.Tags))
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestSubmitMetricsRawDataEmptyArray(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`rawmetrics.submit_raw_metrics_data(None, "checkid", [])`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}

	if len(_rawMetric) != 0 {
		t.Fatalf("Unexpected metric data value: %v", _rawMetric)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

// TODO: Test inner dict
func TestSubmitMetricsRawDataNoArray(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`rawmetrics.submit_raw_metrics_data(None, "checkid", "I should be a array of dict")`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "TypeError: metric data must be a array of dict" {
		t.Errorf("wrong printed value: '%s'", out)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

// TODO
// 	func TestSubmitEventCannotBeSerialized(t *testing.T) {
// 		// Reset memory counters
// 		helpers.ResetMemoryStats()
//
// 		out, err := run(`telemetry.submit_topology_event(None, "checkid", {object(): object()} )`)
//
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		// keys must be a string
// 		if !strings.Contains(out, "keys must be") {
// 			t.Errorf("Unexpected printed value: '%s'", out)
// 		}
// 		if len(_data) != 0 {
// 			t.Fatalf("Unexpected topology event data value: %s", _data)
// 		}
//
// 		// Check for leaks
// 		helpers.AssertMemoryUsage(t)
// 	}
