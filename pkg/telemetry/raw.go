// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package telemetry

import (
	"encoding/json"
	"fmt"
)

// RawMetrics is a batch of raw metrics data
type RawMetrics struct {
	Data []RawMetricsCheckData `json:"check_states"`
}

// RawMetricsCheckData single payload structure
type RawMetricsCheckData struct {
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
func (r RawMetricsCheckData) ConvertToIntakeMetric() []interface{} {
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

// JSONString returns a JSON string of the Component
func (c RawMetricsCheckData) JSONString() string {
	b, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
