package test

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"time"
)

// STSTestCheck is a test implementation of the Check interface
type STSTestCheck struct {
	Name string
}

// String returns the Check Name as a string
func (c *STSTestCheck) String() string { return c.Name }

// Version returns a empty string
func (c *STSTestCheck) Version() string { return "" }

// ConfigSource returns test-config-source
func (c *STSTestCheck) ConfigSource() string { return "test-config-source" }

// Stop is a noop
func (c *STSTestCheck) Stop() {}

// Configure returns nil, noop
func (c *STSTestCheck) Configure(integration.Data, integration.Data, string) error { return nil }

// Interval returns 1
func (c *STSTestCheck) Interval() time.Duration { return 1 }

// Run returns nil
func (c *STSTestCheck) Run() error { return nil }

// ID returns the string as a Check.ID
func (c *STSTestCheck) ID() checkid.ID { return checkid.ID(c.String()) }

// GetWarnings returns an empty []error
func (c *STSTestCheck) GetWarnings() []error { return []error{} }

// GetMetricStats returns an empty map
func (c *STSTestCheck) GetMetricStats() (map[string]int64, error) { return make(map[string]int64), nil }

// IsTelemetryEnabled false for STSTestCheck
func (c *STSTestCheck) IsTelemetryEnabled() bool { return false }