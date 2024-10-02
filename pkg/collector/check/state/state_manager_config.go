package state

import (
	"github.com/DataDog/datadog-agent/pkg/config"
	"time"
)

// Config contains all the configuration for the CheckState
type Config struct {
	StateRootPath           string
	CacheExpirationDuration time.Duration
	CachePurgeDuration      time.Duration
}

// GetStateConfig returns the configuration for the CheckState
func GetStateConfig() Config {
	return Config{
		StateRootPath:           config.Datadog.GetString("check_state_root_path"),
		CacheExpirationDuration: config.Datadog.GetDuration("check_state_expiration_duration"),
		CachePurgeDuration:      config.Datadog.GetDuration("check_state_purge_duration"),
	}
}
