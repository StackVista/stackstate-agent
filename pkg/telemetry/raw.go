// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package telemetry

import (
	"encoding/json"
	"fmt"
)

type RawMetrics struct {
	CheckStates		[]RawMetricsCheckData	`json:"check_states"`
}

type RawMetricsPayload struct {
	Data   RawMetricsCheckData
}

type RawMetricsCheckData struct {
	Name  		string  `json:"name,omitempty"`
	Timestamp 	int64  `json:"timestamp,omitempty"`
	HostName  	string  `json:"host_name,omitempty"`
	Value 		float64  `json:"value,omitempty"`
	Tags      	[]string `json:"tags,omitempty"`
}

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
