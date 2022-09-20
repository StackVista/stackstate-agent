package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
)

// NonTransactionalCheckHandler is a wrapper for check that have no register handler.
type NonTransactionalCheckHandler struct {
	CheckHandlerBase
}

// Name returns NonTransactionalCheckHandler for the non-transactional check handler
func (ch *NonTransactionalCheckHandler) Name() string {
	return "NonTransactionalCheckHandler"
}

// MakeNonTransactionalCheckHandler returns an instance of CheckHandler which functions as a fallback.
func MakeNonTransactionalCheckHandler(check CheckIdentifier, config, initConfig integration.Data) CheckHandler {
	return &NonTransactionalCheckHandler{
		CheckHandlerBase: CheckHandlerBase{
			CheckIdentifier: check,
			config:          config,
			initConfig:      initConfig,
		},
	}
}
