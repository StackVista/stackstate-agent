package check

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
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
	return
}

// GetCheckIdentifier ...
func (ch *CheckNoHandler) GetCheckIdentifier() CheckIdentifier {
	return nil
}

// GetConfig ...
func (ch *CheckNoHandler) GetConfig() (integration.Data, integration.Data) {
	return nil, nil
}
