// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package telemetry

import (
	"encoding/json"
	"fmt"
)

// Metrics is a container structure for a list of raw metric values. This allows us to set Metrics of a batch payload as
// a pointer and append more metrics to the structure
type Metrics struct {
	Values []RawMetrics
}

// RawMetrics single payload structure
type RawMetrics struct {
	Name      string   `json:"name,omitempty"`
	Timestamp int64    `json:"timestamp,omitempty"`
	HostName  string   `json:"host_name,omitempty"`
	Value     float64  `json:"value,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// RawMetricsMetaData payload containing meta data for the metric
type RawMetricsMetaData struct {
	Hostname string   `json:"hostname,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Type     string   `json:"type,omitempty"`
}

// ConvertToIntakeMetric Converts RawMetricsCheckData struct to an older v1 metrics structure
func (r RawMetrics) ConvertToIntakeMetric() []interface{} {
	data := []interface{}{
		r.Name,
		r.Timestamp,
		r.Value,
		RawMetricsMetaData{
			Hostname: r.HostName,
			Type:     "raw",
			Tags:     r.Tags,
		},
	}
	return data
}

// IntakeMetricJSON Converts RawMetricsCheckData struct to an older v1 metrics structure, parses it to JSON and returns
// it as a interface. This is only used in batcher test assertions.
func (r RawMetrics) IntakeMetricJSON() (jsonObject []interface{}) {
	jsonString, _ := json.Marshal(r.ConvertToIntakeMetric())
	_ = json.Unmarshal(jsonString, &jsonObject)
	return jsonObject
}

// JSONString returns a JSON string of the Component
func (r RawMetrics) JSONString() string {
	b, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
