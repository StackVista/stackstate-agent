package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
)

// CheckNoReloader is a implementation of the CheckLoader interface that does a noop on ReloadCheck
type CheckNoReloader struct{}

// ReloadCheck returns nil
func (n CheckNoReloader) ReloadCheck(check.ID, integration.Data, integration.Data, string) error {
	return nil
}
