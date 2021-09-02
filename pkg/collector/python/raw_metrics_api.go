// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

// +build python

package python

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

/*
#cgo !windows LDFLAGS: -ldatadog-agent-rtloader -ldl
#cgo windows LDFLAGS: -ldatadog-agent-rtloader -lstdc++ -static

#include "datadog_agent_rtloader.h"
#include "rtloader_mem.h"
*/
import "C"

// NOTE
// Beware that any changes made here MUST be reflected also in the test implementation
// rtloader/test/raw_metrics/raw_metrics.go

// SubmitRawMetricsData
//export SubmitRawMetricsData
func SubmitRawMetricsData(id *C.char, data *C.char) {
	goCheckID := C.GoString(id)
	rawMetricsPayload := C.GoString(data)
	metricsPayload := metrics.RawMetricsPayload{}
	err := json.Unmarshal([]byte(rawMetricsPayload), &metricsPayload)

	if err == nil {
		if len(metricsPayload.Data) != 0 {
			batcher.GetBatcher().SubmitRawMetricsData(check.ID(goCheckID), metricsPayload.Data)
		} else {
			_ = log.Errorf("Empty json submitted to as check data, this is not allowed, data will not be forwarded.")
		}
	} else {
		_ = log.Errorf("Empty raw metrics data event not sent. Raw: %v, Json: %v, Error: %v", rawMetricsPayload,
			metricsPayload.JSONString(), err)
	}
}
