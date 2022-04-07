package check

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"time"
)

// FIXTURE
type TestCheck struct {
	Name string
}

func (c *TestCheck) String() string                                             { return c.Name }
func (c *TestCheck) Version() string                                            { return "" }
func (c *TestCheck) ConfigSource() string                                       { return "test-config-source" }
func (c *TestCheck) Stop()                                                      {}
func (c *TestCheck) Configure(integration.Data, integration.Data, string) error { return nil }
func (c *TestCheck) Interval() time.Duration                                    { return 1 }
func (c *TestCheck) Run() error                                                 { return nil }
func (c *TestCheck) ID() ID                                                     { return ID(c.String()) }
func (c *TestCheck) GetWarnings() []error                                       { return []error{} }
func (c *TestCheck) GetMetricStats() (map[string]int64, error)                  { return make(map[string]int64), nil }
func (c *TestCheck) IsTelemetryEnabled() bool                                   { return false }

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
