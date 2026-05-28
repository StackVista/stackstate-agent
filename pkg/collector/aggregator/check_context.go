// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package aggregator

import (
	"errors"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"

	tagger "github.com/DataDog/datadog-agent/comp/core/tagger/def"
	"github.com/DataDog/datadog-agent/comp/core/tagger/types"
	workloadfilter "github.com/DataDog/datadog-agent/comp/core/workloadfilter/def"
	integrations "github.com/DataDog/datadog-agent/comp/logs/integrations/def"
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/DataDog/datadog-agent/pkg/util/option"
)

var checkCtx *CheckContext
var checkContextMutex = sync.Mutex{}

// CheckContext stores the global context required by Go methods like SubmitMetric.
// Doing so allow to have a single global state instead of having one
// per dependency used inside SubmitMetric like methods.
type CheckContext struct {
	senderManager sender.SenderManager
	checkManager  handler.CheckManager
	logReceiver   option.Option[integrations.Component]
	tagger        tagger.Component
	filter        workloadfilter.FilterBundle
}

func (cc *CheckContext) Tag(entityID types.EntityID, cardinality types.TagCardinality) ([]string, error) {
	return cc.tagger.Tag(entityID, cardinality)
}

func (cc *CheckContext) GetLogReceiver() (integrations.Component, bool) {
	return cc.logReceiver.Get()
}

func (cc *CheckContext) IsExcluded(container *workloadfilter.Container) bool {
	return cc.filter.IsExcluded(container)
}

// GetCheckContext retrives the current context
func GetCheckContext() (*CheckContext, error) {
	checkContextMutex.Lock()
	defer checkContextMutex.Unlock()

	if checkCtx == nil {
		return nil, errors.New("Python check context was not set")
	}
	return checkCtx, nil
}

// InitializeCheckContext creates the context that can be later used for storing/retrieving checks context for submit functions
func InitializeCheckContext(senderManager sender.SenderManager, checkManager handler.CheckManager, logReceiver option.Option[integrations.Component], tagger tagger.Component, filterStore workloadfilter.Component) {
	checkContextMutex.Lock()
	if checkCtx == nil {
		checkCtx = &CheckContext{
			senderManager: senderManager,
			checkManager:  checkManager,
			logReceiver:   logReceiver,
			tagger:        tagger,
			filter:        filterStore.GetContainerSharedMetricFilters(),
		}

		if _, ok := logReceiver.Get(); !ok {
			// [sts] Downgraded from Warn to Info: regulated customers (military, banks)
			// raise tickets on WARN entries. STS doesn't wire up the integration-logs
			// pipeline by default, so this fires at every agent start — it's expected
			// configuration state, not a problem.
			log.Info("Log receiver not provided. Logs from integrations will not be collected.")
		}
	}

	checkContextMutex.Unlock()
}

// Testing utilities - Made test execution mutexed to avoid race conditions during testing.
var testMutex = sync.Mutex{}

func withLockedCheckContext(senderManager sender.SenderManager, checkManager handler.CheckManager, logReceiver option.Option[integrations.Component], tagger tagger.Component, filterStore workloadfilter.Component) {
	testMutex.Lock()
	// [sts] Defensive cleanup: if a previous test left a context behind (e.g.,
	// test panicked between ScopeInitCheckContext returning and `defer release()`
	// being registered), reset state instead of panicking — a panic here while
	// holding checkContextMutex would deadlock subsequent test runs of this
	// binary. Adopting the cleanup behavior also removes the need for the
	// pre-init defensive call in ScopeInitCheckContext.
	checkContextMutex.Lock()
	if checkCtx != nil {
		checkCtx.checkManager.Stop()
		checkCtx = nil
	}
	checkContextMutex.Unlock()
	InitializeCheckContext(senderManager, checkManager, logReceiver, tagger, filterStore)
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
