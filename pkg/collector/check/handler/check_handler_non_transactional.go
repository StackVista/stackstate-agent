package handler

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/collector/check/state"
)

// NonTransactionalCheckHandler is a wrapper for check that have no register handler.
type NonTransactionalCheckHandler struct {
	CheckHandlerBase
	manager CheckManager
	batcher batcher.Component
}

// Name returns NonTransactionalCheckHandler for the non-transactional check handler
func (ch *NonTransactionalCheckHandler) Name() string {
	return "NonTransactionalCheckHandler"
}

// MakeNonTransactionalCheckHandler returns an instance of CheckHandler which functions as a fallback.
func MakeNonTransactionalCheckHandler(manager CheckManager, batcher batcher.Component, stateManager state.CheckStateAPI, check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	return &NonTransactionalCheckHandler{
		CheckHandlerBase: CheckHandlerBase{
			CheckIdentifier: check,
			config:          config,
			initConfig:      initConfig,
			stateManager:    stateManager,
		},
		manager: manager,
		batcher: batcher,
	}
}
