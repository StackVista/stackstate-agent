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

func MakeCheckNoHandler() CheckHandler {
	return &checkHandler{CheckHandlerBase{
		Batcher: batcher.GetBatcher(),
	}}
}

// ReloadCheck ...
func (ch *CheckNoHandler) ReloadCheck() {
	_ = log.Warnf("ReloadCheck called on CheckNoHandler. This should never happen.")
}

// GetCheckIdentifier ...
func (ch *CheckNoHandler) GetCheckIdentifier() CheckIdentifier {
	return nil
}

// GetConfig ...
func (ch *CheckNoHandler) GetConfig() (integration.Data, integration.Data) {
	return []byte{}, []byte{}
}
