// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at StackState (https://www.stackstate.com).
// Copyright 2021 StackState

//go:build python
// +build python

package python

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/checkmanager"
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

// SubmitStartTransaction starts a transaction
//export SubmitStartTransaction
func SubmitStartTransaction(id *C.char) {
	goCheckID := C.GoString(id)
	checkmanager.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SubmitStartTransaction()
}

// SubmitStopTransaction stops a transaction
//export SubmitStopTransaction
func SubmitStopTransaction(id *C.char) {
	goCheckID := C.GoString(id)
	checkmanager.GetCheckManager().GetCheckHandler(check.ID(goCheckID)).SubmitStopTransaction()
}
