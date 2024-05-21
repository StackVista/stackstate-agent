// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

//go:build python
// +build python

package python

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
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
// rtloader/test/telemetry/telemetry.go

// SubmitTopologyEvent is the method exposed to Python scripts to submit topology event
//export SubmitTopologyEvent
func SubmitTopologyEvent(id *C.char, data *C.char) {
	goCheckID := C.GoString(id)

	var err error

	var topologyEvent metrics.Event
	rawEvent := C.GoString(data)
	err = json.Unmarshal([]byte(rawEvent), &topologyEvent)

	if err == nil {
		// [sts] send events via die check handler
		handler.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SubmitEvent(topologyEvent)
	} else {
		_ = log.Errorf("Empty topology event not sent. Raw: %v, Json: %v, Error: %v", rawEvent,
			topologyEvent.String(), err)
	}
}

// SubmitRawMetricsData
//export SubmitRawMetricsData
func SubmitRawMetricsData(checkID *C.char, name *C.char, value C.float, tags **C.char, hostname *C.char, timestamp C.longlong) {
	goCheckID := C.GoString(checkID)
	rawName := C.GoString(name)
	rawHostname := C.GoString(hostname)
	rawValue := float64(value)
	rawTimestamp := int64(timestamp)
	rawTags := cStringArrayToSlice(tags)

	handler.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SubmitRawMetricsData(telemetry.RawMetrics{
		Name:      rawName,
		Timestamp: rawTimestamp,
		HostName:  rawHostname,
		Value:     rawValue,
		Tags:      rawTags,
	})
}
