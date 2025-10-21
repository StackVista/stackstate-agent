// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

//go:build python

package python

import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
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
// rtloader/test/transaction/transaction.go

// StartTransaction starts a transaction
//
//export StartTransaction
func StartTransaction(id *C.char) {
	goCheckID := C.GoString(id)

	checkContext, err := getCheckContext()
	if err != nil {
		log.Errorf("Python check context: %v", err)
		return
	}

	checkContext.checkManager.GetCheckHandler(checkid.ID(goCheckID)).StartTransaction()
}

// StopTransaction stops a transaction
//
//export StopTransaction
func StopTransaction(id *C.char) {
	goCheckID := C.GoString(id)

	checkContext, err := getCheckContext()
	if err != nil {
		log.Errorf("Python check context: %v", err)
		return
	}

	checkContext.checkManager.GetCheckHandler(checkid.ID(goCheckID)).StopTransaction()
}

// DiscardTransaction cancels a transaction
//
//export DiscardTransaction
func DiscardTransaction(id *C.char, reason *C.char) {
	goCheckID := C.GoString(id)

	checkContext, err := getCheckContext()
	if err != nil {
		log.Errorf("Python check context: %v", err)
		return
	}

	goReason := C.GoString(reason)

	checkContext.checkManager.GetCheckHandler(checkid.ID(goCheckID)).DiscardTransaction(goReason)
}

// SetTransactionState sets a state for a transaction
//
//export SetTransactionState
func SetTransactionState(id *C.char, key *C.char, state *C.char) {
	goCheckID := C.GoString(id)

	checkContext, err := getCheckContext()
	if err != nil {
		log.Errorf("Python check context: %v", err)
		return
	}

	keyValue := C.GoString(key)
	stateValue := C.GoString(state)

	checkContext.checkManager.GetCheckHandler(checkid.ID(goCheckID)).SetTransactionState(keyValue, stateValue)
}
