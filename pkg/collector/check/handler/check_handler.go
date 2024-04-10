package handler

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
)

// CheckIdentifier encapsulates all the functionality needed to describe and configure an agent check.
type CheckIdentifier interface {
	String() string       // provide a printable version of the check name
	ID() check.ID         // provide a unique identifier for every check instance
	ConfigSource() string // return the configuration source of the check
}

// CheckHandler represents a wrapper around an Agent Check that allows us track data produced by an agent check, as well
// as handle transactions for it.
type CheckHandler interface {
	CheckIdentifier
	CheckAPI
	Name() string
	GetConfig() (config, initConfig integration.Data)
}
