package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckNoHandler ...
type CheckNoHandler struct {
	CheckHandlerBase
}

// MakeCheckNoHandler returns an instance of CheckHandler which functions as a fallback
func MakeCheckNoHandler() CheckHandler {
	return &checkHandler{CheckHandlerBase{
		Batcher: batcher.GetBatcher(),
	}}
}

// ReloadCheck is the CheckNoHandler implementation which is a no-op
func (ch *CheckNoHandler) ReloadCheck() {
	_ = log.Warnf("ReloadCheck called on CheckNoHandler. This should never happen.")
}

// GetCheckIdentifier is the CheckNoHandler implementation which just returns nil. This should never be called.
func (ch *CheckNoHandler) GetCheckIdentifier() CheckIdentifier {
	_ = log.Warnf("GetCheckIdentifier called on CheckNoHandler. This should never happen.")
	return nil
}

// GetConfig is the CheckNoHandler implementation which just returns nil. This should never be called.
func (ch *CheckNoHandler) GetConfig() (integration.Data, integration.Data) {
	_ = log.Warnf("GetConfig called on CheckNoHandler. This should never happen.")
	return nil, nil
}
