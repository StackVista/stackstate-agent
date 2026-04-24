package state

import (
	"time"

	pkgconfigsetup "github.com/DataDog/datadog-agent/pkg/config/setup"
)

// Config contains all the configuration for the CheckState
type Config struct {
	StateRootPath           string
	CacheExpirationDuration time.Duration
	CachePurgeDuration      time.Duration
}

// GetStateConfig returns the configuration for the CheckState
func GetStateConfig() Config {
	stateRootPath := pkgconfigsetup.Datadog().GetString("check_state_root_path")
	if stateRootPath == "" {
		stateRootPath = pkgconfigsetup.Datadog().GetString("run_path")
	}
	return Config{
		StateRootPath:           stateRootPath,
		CacheExpirationDuration: pkgconfigsetup.Datadog().GetDuration("check_state_expiration_duration"),
		CachePurgeDuration:      pkgconfigsetup.Datadog().GetDuration("check_state_purge_duration"),
	}
}
