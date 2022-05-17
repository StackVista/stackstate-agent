package check

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"time"
)

// FIXTURE
type STSTestCheck struct {
	Name string
}

func (c *STSTestCheck) String() string                                             { return c.Name }
func (c *STSTestCheck) Version() string                                            { return "" }
func (c *STSTestCheck) ConfigSource() string                                       { return "test-config-source" }
func (c *STSTestCheck) Stop()                                                      {}
func (c *STSTestCheck) Configure(integration.Data, integration.Data, string) error { return nil }
func (c *STSTestCheck) Interval() time.Duration                                    { return 1 }
func (c *STSTestCheck) Run() error                                                 { return nil }
func (c *STSTestCheck) ID() ID                                                     { return ID(c.String()) }
func (c *STSTestCheck) GetWarnings() []error                                       { return []error{} }
func (c *STSTestCheck) GetMetricStats() (map[string]int64, error)                  { return make(map[string]int64), nil }
func (c *STSTestCheck) IsTelemetryEnabled() bool                                   { return false }

// TestCheckReloader is an implementation of the CheckLoader
type TestCheckReloader struct {
	Reloaded int
}

// GetReloaded returns the integer representing the amount of time reloaded was called
func (tcr *TestCheckReloader) GetReloaded() int {
	return tcr.Reloaded
}

// ReloadCheck increments the reloaded integer
func (tcr *TestCheckReloader) ReloadCheck(ID, integration.Data, integration.Data, string) error {
	tcr.Reloaded = tcr.Reloaded + 1
	return nil
}
