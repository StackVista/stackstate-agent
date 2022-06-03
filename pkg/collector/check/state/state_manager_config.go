package state

import (
	"github.com/StackVista/stackstate-agent/pkg/config"
	"time"
)

// StateConfig contains all the configuration for the CheckState
type StateConfig struct {
	StateRootPath           string
	CacheExpirationDuration time.Duration
	CachePurgeDuration      time.Duration
}

// GetStateConfig returns the configuration for the CheckState
func GetStateConfig() StateConfig {
	return StateConfig{
		StateRootPath:           config.Datadog.GetString("check_state_root_path"),
		CacheExpirationDuration: config.Datadog.GetDuration("check_state_expiration_duration"),
		CachePurgeDuration:      config.Datadog.GetDuration("check_state_purge_duration"),
	}
}
