package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckHandlerBase forms the base of the transactional and non-transactional check handler
type CheckHandlerBase struct {
	CheckIdentifier
	CheckReloader
	config, initConfig integration.Data
}

// Reload uses the Check Reloader (Collector) to reload the check: pkg/collector/collector.go:126
func (ch *CheckHandlerBase) Reload() {
	config, initConfig := ch.GetConfig()
	if err := ch.CheckReloader.ReloadCheck(ch.ID(), config, initConfig, ch.ConfigSource()); err != nil {
		_ = log.Errorf("could not reload check %s", string(ch.ID()))
	}
}

// GetConfig returns the config and the init config of the check
func (ch *CheckHandlerBase) GetConfig() (integration.Data, integration.Data) {
	return ch.config, ch.initConfig
}

// GetCheckReloader returns the configured CheckReloader.
func (ch *CheckHandlerBase) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}
