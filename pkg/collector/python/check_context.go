// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build python

package python

import (
	"errors"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
)

var checkCtx *checkContext
var checkContextMutex = sync.Mutex{}

// As it is difficult to pass Go context to Go methods like SubmitMetric,
// checkContext stores the global context required by these functions.
// Doing so allow to have a single global state instead of having one
// per dependency used inside SubmitMetric like methods.
type checkContext struct {
	senderManager sender.SenderManager
	checkManager  handler.CheckManager
}

func getCheckContext() (*checkContext, error) {
	checkContextMutex.Lock()
	defer checkContextMutex.Unlock()

	if checkCtx == nil {
		return nil, errors.New("Python check context was not set")
	}
	return checkCtx, nil
}

func initializeCheckContext(senderManager sender.SenderManager, checkManager handler.CheckManager) {
	checkContextMutex.Lock()
	if checkCtx == nil {
		checkCtx = &checkContext{senderManager: senderManager, checkManager: checkManager}
	}
	checkContextMutex.Unlock()
}

// Testing utilities
var testMutex = sync.Mutex{}

func withLockedCheckContext(senderManager sender.SenderManager, checkManager handler.CheckManager) {
	testMutex.Lock()
	checkContextMutex.Lock()
	if checkCtx != nil {
		panic("CheckContext was left initialized")
	}
	checkContextMutex.Unlock()
	initializeCheckContext(senderManager, checkManager)
}

func releaseCheckContext() {
	checkContextMutex.Lock()
	if checkCtx != nil {
		checkCtx.checkManager.Stop()
	}

	checkCtx = nil
	checkContextMutex.Unlock()
	testMutex.Unlock()
}