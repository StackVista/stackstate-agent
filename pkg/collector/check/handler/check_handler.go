package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
)

// CheckReloader is a interface wrapper around the Collector which controls all checks.
// TODO BUG: 2022/08/01, Description of bug below
// The ReloadCheck functionality works as intended but we noticed that when you run the reload check paired with
// within a Check cycle, for example, within a Transaction Discard event if you call ReloadCheck you will get GoRoutine
// Race conditions, This needs to be investigated and fixed if this function is used.
// The functionality is left here so that we do not have to rewrite everything and only fix the bug if a reloadcheck
// is require.
type CheckReloader interface {
	ReloadCheck(id check.ID, config, initConfig integration.Data, newSource string) error
}

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
	GetCheckReloader() CheckReloader
	Reload()
}
