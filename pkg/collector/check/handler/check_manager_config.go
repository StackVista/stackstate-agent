package handler

import "github.com/StackVista/stackstate-agent/pkg/config"

// Config contains all the configuration values for the check manager
type Config struct {
	CheckTransactionalityEnabled bool
}

// GetCheckManagerConfig returns the configuration for the checkmanager
func GetCheckManagerConfig() Config {
	return Config{
		CheckTransactionalityEnabled: config.Datadog.GetBool("check_transactionality_enabled"),
	}
}
