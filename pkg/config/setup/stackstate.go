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

	// [sts] STS-specific kubernetes collector toggles. Registered in 7.71.2's
	// pkg/config/setup/config.go and migrated here when DD 7.78.x split per-component
	// init functions. Read by pkg/collector/corechecks/cluster/kubeapi/{kubernetes_topology_config,kubernetes_metrics}.go
	// and asserted by pkg/config/legacy/kubernetes_test.go TestConvertKubernetes.
	config.BindEnvAndSetDefault("collect_kubernetes_metrics", false)
	config.BindEnvAndSetDefault("collect_kubernetes_topology", false)
	config.BindEnvAndSetDefault("collect_kubernetes_timeout", 10)
	config.BindEnvAndSetDefault("configmap_max_datasize", 0)
	config.BindEnvAndSetDefault("kubernetes_csi_pv_mapper_enabled", false)

	// [sts] Tolerate kubelet TLS verification failures and insecure transports.
	// STS clusters often run with self-signed kubelet certs; the DD default
	// (false) breaks discovery on those clusters. Read by
	// pkg/util/kubernetes/kubelet/kubelet_client.go.
	config.BindEnvAndSetDefault("kubelet_fallback_to_unverified_tls", true)
	config.BindEnvAndSetDefault("kubelet_fallback_to_insecure", true)
}

// GetMaxCapacity returns the maximum amount of elements per batch for the transactionbatcher.
// [sts]
func GetMaxCapacity(config pkgconfigmodel.Reader) int {
	if config.IsSet("batcher_capacity") {
		return config.GetInt("batcher_capacity")
	}
	return DefaultBatcherBufferSize
}

// GetTxManagerConfig returns the transaction manager configuration: buffer size, ticker interval,
// timeout duration, eviction duration.
// [sts]
func GetTxManagerConfig(config pkgconfigmodel.Reader) (int, time.Duration, time.Duration, time.Duration) {
	txBufferSize := Datadog().GetInt("transaction_manager_channel_buffer_size")
	txTickerInterval := time.Second * time.Duration(config.GetInt("transaction_ticket_interval_seconds"))
	txTimeoutDuration := time.Second * time.Duration(config.GetInt("transaction_timeout_duration_seconds"))
	txEvictionDuration := time.Second * time.Duration(config.GetInt("transaction_eviction_duration_seconds"))
	return txBufferSize, txTickerInterval, txTimeoutDuration, txEvictionDuration
}
