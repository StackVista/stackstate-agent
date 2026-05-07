// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package setup contains StackState-specific config defaults applied via the
// stackstate() initializer registered in commonConfigComponents.
package setup

import (
	"time"

	pkgconfigmodel "github.com/DataDog/datadog-agent/pkg/config/model"
)

const (
	// DefaultBatcherBufferSize sets the default buffer size of the batcher to 10000
	// [sts]
	DefaultBatcherBufferSize = 10000

	// DefaultTxManagerChannelBufferSize is the concurrent transactions before the tx manager begins backpressure
	// [sts] transaction manager
	DefaultTxManagerChannelBufferSize = 100
	// DefaultTxManagerTimeoutDurationSeconds is the amount of time before a transaction is marked as stale, 5 minutes by default
	DefaultTxManagerTimeoutDurationSeconds = 60 * 5
	// DefaultTxManagerEvictionDurationSeconds is the amount of time before a transaction is evicted and rolled back, 10 minutes by default
	DefaultTxManagerEvictionDurationSeconds = 60 * 10
	// DefaultTxManagerTickerIntervalSeconds is the ticker interval to mark transactions as stale / timeout.
	DefaultTxManagerTickerIntervalSeconds = 30

	// DefaultCheckStateExpirationDuration is the amount of time before an element is expired from the Check State cache, 10 minutes by default
	// [sts]
	DefaultCheckStateExpirationDuration = 10 * time.Minute
	// DefaultCheckStatePurgeDuration is the amount of time before an element is removed from the Check State cache, 10 minutes by default
	DefaultCheckStatePurgeDuration = 10 * time.Minute
)

// stackstate registers StackState-specific config defaults. Registered in
// commonConfigComponents in config.go.
func stackstate(config pkgconfigmodel.Setup) {
	// [sts] skip datadog functionality
	config.BindEnvAndSetDefault("skip_leader_election", true)
	// [sts] bind env for skip_validate_clustername, default is set in the config_template.yaml to avoid test failures.
	config.BindEnv("skip_validate_clustername") //nolint:errcheck

	// [sts] batcher environment variables
	config.BindEnvAndSetDefault("batcher_capacity", DefaultBatcherBufferSize)

	// [sts] transactional environment variables
	config.BindEnvAndSetDefault("transaction_manager_channel_buffer_size", DefaultTxManagerChannelBufferSize)
	config.BindEnvAndSetDefault("transaction_timeout_duration_seconds", DefaultTxManagerTimeoutDurationSeconds)
	config.BindEnvAndSetDefault("transaction_eviction_duration_seconds", DefaultTxManagerEvictionDurationSeconds)
	config.BindEnvAndSetDefault("transaction_ticket_interval_seconds", DefaultTxManagerTickerIntervalSeconds)

	// [sts] check state manager environment variable
	config.BindEnvAndSetDefault("check_state_root_path", "")
	config.BindEnvAndSetDefault("check_state_expiration_duration", DefaultCheckStateExpirationDuration)
	config.BindEnvAndSetDefault("check_state_purge_duration", DefaultCheckStatePurgeDuration)

	// [sts] retryable http client environment variables
	config.BindEnvAndSetDefault("transactional_forwarder_retry_min", 1*time.Second)
	config.BindEnvAndSetDefault("transactional_forwarder_retry_max", 10*time.Second)

	config.BindEnvAndSetDefault("skip_hostname_validation", false) // sts

	// [sts] disable periodic connectivity checker — STS receiver does not support all DD endpoints,
	// causing 404s in receiver logs every 10 minutes
	config.BindEnvAndSetDefault("connectivity_checker.enabled", false)
}
