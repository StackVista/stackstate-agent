package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckNoHandler ...
type CheckNoHandler struct {
	CheckHandlerBase
}

// MakeCheckNoHandler returns an instance of CheckHandler which functions as a fallback
func MakeCheckNoHandler() CheckHandler {
	return &CheckNoHandler{CheckHandlerBase: CheckHandlerBase{}}
}

// StartTransaction ...
func (ch *CheckNoHandler) StartTransaction(check.ID, string) {
	_ = log.Warnf("StartTransaction called on CheckNoHandler. This should never happen.")
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

func (ch *CheckNoHandler) GetBatcher() batcher.Batcher {
	return batcher.GetBatcher()
}

func (ch *CheckNoHandler) GetCheckReloader() CheckReloader {
	return NoCheckReloader{}
}

type NoCheckReloader struct{}

func (n NoCheckReloader) ReloadCheck(id check.ID, config, initConfig integration.Data, newSource string) error {
	return nil
}
