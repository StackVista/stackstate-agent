package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
)

// CheckHandlerBase forms the base of the transactional and non-transactional check handler
type CheckHandlerBase struct {
	CheckIdentifier
	config, initConfig integration.Data
}

// GetConfig returns the config and the init config of the check
func (ch *CheckHandlerBase) GetConfig() (integration.Data, integration.Data) {
	return ch.config, ch.initConfig
}
