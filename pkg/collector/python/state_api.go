// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

//go:build python
// +build python

package python

/*
#cgo !windows LDFLAGS: -ldatadog-agent-rtloader -ldl
#cgo windows LDFLAGS: -ldatadog-agent-rtloader -lstdc++ -static

#include "datadog_agent_rtloader.h"
#include "rtloader_mem.h"
*/
import "C"
import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
)

// NOTE
// Beware that any changes made here MUST be reflected also in the test implementation
// rtloader/test/state/state.go

// SetState set the current state
//export SetState
func SetState(id *C.char, key *C.char, state *C.char) {
	goCheckID := C.GoString(id)
	stateKey := C.GoString(key)
	stateValue := C.GoString(state)

	handler.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SetState(stateKey, stateValue)
}

// GetState get the current state
//export GetState
func GetState(id *C.char, key *C.char) *C.char {
	goCheckID := C.GoString(id)
	stateKey := C.GoString(key)

	getStateResult := handler.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).GetState(stateKey)

	return C.CString(getStateResult)
}
