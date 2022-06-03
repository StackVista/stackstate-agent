package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// NonTransactionalCheckHandler is a wrapper for check that have no register handler.
type NonTransactionalCheckHandler struct {
	CheckIdentifier
	CheckReloader CheckReloader
}

// MakeNonTransactionalCheckHandler returns an instance of CheckHandler which functions as a fallback.
func MakeNonTransactionalCheckHandler(check CheckIdentifier, cr CheckReloader) CheckHandler {
	return &NonTransactionalCheckHandler{
		CheckIdentifier: check,
		CheckReloader:   cr,
	}
}

// Reload is a wrapper around the CheckReloader reload function
func (ch *NonTransactionalCheckHandler) Reload() {
	config, initConfig := ch.GetConfig()
	_ = ch.CheckReloader.ReloadCheck(ch.ID(), config, initConfig, ch.ConfigSource())
}

// GetConfig is the NonTransactionalCheckHandler implementation which just returns nil. This should never be called.
func (ch *NonTransactionalCheckHandler) GetConfig() (integration.Data, integration.Data) {
	_ = log.Warnf("GetConfig called on NonTransactionalCheckHandler. This should never happen.")
	return nil, nil
}

// GetBatcher returns the global batcher instance (non-transactional)
func (ch *NonTransactionalCheckHandler) GetBatcher() batcher.Batcher {
	return batcher.GetBatcher()
}

// GetCheckReloader returns the configured CheckReloader
func (ch *NonTransactionalCheckHandler) GetCheckReloader() CheckReloader {
	return ch.CheckReloader
}

// NewCheckIdentifier returns a IDOnlyCheckIdentifier that implements the CheckIdentifier interface for a given check ID.
func NewCheckIdentifier(checkID check.ID) CheckIdentifier {
	return &IDOnlyCheckIdentifier{checkID: checkID}
}

// IDOnlyCheckIdentifier is used to create a CheckIdentifier when only the checkID is present
type IDOnlyCheckIdentifier struct {
	checkID check.ID
}

// String returns the IDOnlyCheckIdentifier checkID as a string
func (idCI *IDOnlyCheckIdentifier) String() string {
	return string(idCI.checkID)
}

// ID returns the IDOnlyCheckIdentifier checkID
func (idCI *IDOnlyCheckIdentifier) ID() check.ID {
	return idCI.checkID
}

// ConfigSource returns an empty string for the IDOnlyCheckIdentifier
func (*IDOnlyCheckIdentifier) ConfigSource() string {
	return ""
}

// NoCheckReloader is a implementation of the CheckLoader interface that does a noop on ReloadCheck
type NoCheckReloader struct{}

// ReloadCheck returns nil
func (n NoCheckReloader) ReloadCheck(check.ID, integration.Data, integration.Data, string) error {
	return nil
}
