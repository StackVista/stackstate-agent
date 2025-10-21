package handler

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check/state"
)

// CheckHandlerBase forms the base of the transactional and non-transactional check handler
type CheckHandlerBase struct {
	CheckIdentifier
	config, initConfig integration.Data
	stateManager       state.CheckStateAPI
}

// GetConfig returns the config and the init config of the check
func (ch *CheckHandlerBase) GetConfig() (integration.Data, integration.Data) {
	return ch.config, ch.initConfig
}
