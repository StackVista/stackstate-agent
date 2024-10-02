// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

//go:build python

package python

/*
#cgo !windows LDFLAGS: -ldatadog-agent-rtloader -ldl
#cgo windows LDFLAGS: -ldatadog-agent-rtloader -lstdc++ -static

#include "datadog_agent_rtloader.h"
#include "rtloader_mem.h"
*/
import "C"
import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// NOTE
// Beware that any changes made here MUST be reflected also in the test implementation
// rtloader/test/state/state.go

// SetState set the current state
//
//export SetState
func SetState(id *C.char, key *C.char, state *C.char) {
	goCheckID := C.GoString(id)

	checkContext, err := getCheckContext()
	if err != nil {
		log.Errorf("Python check context: %v", err)
		return
	}

	stateKey := C.GoString(key)
	stateValue := C.GoString(state)

	checkContext.checkManager.GetCheckHandler(checkid.ID(goCheckID)).SetState(stateKey, stateValue)
}

// GetState get the current state
//
//export GetState
func GetState(id *C.char, key *C.char) *C.char {
	goCheckID := C.GoString(id)
	stateKey := C.GoString(key)

	checkContext, err := getCheckContext()
	if err != nil {
		log.Errorf("Python check context: %v", err)
		return nil
	}

	getStateResult := checkContext.checkManager.GetCheckHandler(checkid.ID(goCheckID)).GetState(stateKey)

	return C.CString(getStateResult)
}
