// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package metrics

import (
	"encoding/json"
	"fmt"
)

// TODO: Raw Metrics

type RawMetrics struct {
	Stream        RawMetricsStream       	`json:"stream"`
	CheckStates   []RawMetricsCheckData		`json:"check_states"`
}

type RawMetricsStream struct {
	Urn       string `json:"urn"`
	SubStream string `json:"sub_stream,omitempty"`
}

type RawMetricsCheckData map[string]interface{}

type RawMetricsPayload struct {
	Stream RawMetricsStream
	Data   RawMetricsCheckData
}

// TODO: Make generic with health check

// JSONString returns a JSON string of the Payload
func (p RawMetricsPayload) JSONString() string {
	b, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
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

// GoString prints as string, can also be used in maps
func (i *RawMetricsStream) GoString() string {
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}

// 	type RawMetricsData struct {
// 		Name      string            `json:"name"`
// 		Timestamp int               `json:"timestamp"`
// 		Value     string            `json:"value"`
// 		Hostname  string            `json:"hostname"`
// 		Type      string            `json:"type,omitempty"`
// 		Tags      map[string]string `json:"tags"`
// 	}
