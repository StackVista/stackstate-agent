// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

//go:build python
// +build python

package python

import (
	"encoding/json"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/health"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/log"
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
// rtloader/test/health/health.go

// SubmitHealthCheckData is the method exposed to Python scripts to submit health check data
//
//export SubmitHealthCheckData
func SubmitHealthCheckData(id *C.char, _ *C.health_stream_t, data *C.char) {
	goCheckID := C.GoString(id)
	rawHealthPayload := C.GoString(data)
	healthPayload := health.Payload{}
	err := json.Unmarshal([]byte(rawHealthPayload), &healthPayload)

	if err == nil {
		if !healthPayload.Data.IsEmpty() {
			handler.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SubmitHealthCheckData(healthPayload.Stream, healthPayload.Data)
		} else {
			_ = log.Errorf("Empty json submitted to as check data, this is not allowed, data will not be forwarded.")
		}
	} else {
		_ = log.Errorf("Empty health data event not sent. Raw: %v, Json: %v, Error: %v", rawHealthPayload,
			util.JSONString(healthPayload), err)
	}
}

// SubmitHealthStartSnapshot starts a health snapshot
//
//export SubmitHealthStartSnapshot
func SubmitHealthStartSnapshot(id *C.char, healthStream *C.health_stream_t, expirySeconds C.int, repeatIntervalSeconds C.int) {
	goCheckID := C.GoString(id)
	_stream := convertStream(healthStream)

	handler.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SubmitHealthStartSnapshot(_stream, int(repeatIntervalSeconds), int(expirySeconds))
}

// SubmitHealthStopSnapshot stops a health snapshot
//
//export SubmitHealthStopSnapshot
func SubmitHealthStopSnapshot(id *C.char, healthStream *C.health_stream_t) {
	goCheckID := C.GoString(id)
	_stream := convertStream(healthStream)

	handler.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SubmitHealthStopSnapshot(_stream)
}

func convertStream(healthStream *C.health_stream_t) health.Stream {
	_subStream := C.GoString(healthStream.sub_stream)
	if _subStream == "" {
		return health.Stream{
			Urn: C.GoString(healthStream.urn),
		}
	}
	return health.Stream{
		Urn:       C.GoString(healthStream.urn),
		SubStream: _subStream,
	}
}
