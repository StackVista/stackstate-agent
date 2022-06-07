//go:build !linux
// +build !linux

// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package app

import (
	"errors"

	coreconfig "github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/logs/client"
	"github.com/StackVista/stackstate-agent/pkg/logs/config"
	"github.com/StackVista/stackstate-agent/pkg/logs/restart"
	secagent "github.com/StackVista/stackstate-agent/pkg/security/agent"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

func startRuntimeSecurity(hostname string, endpoints *config.Endpoints, context *client.DestinationsContext, stopper restart.Stopper) (*secagent.RuntimeSecurityAgent, error) {
	enabled := coreconfig.Datadog.GetBool("runtime_security_config.enabled")
	if !enabled {
		log.Info("Datadog runtime security agent disabled by config")
		return nil, nil
	}

	return nil, errors.New("Datadog runtime security agent is only supported on Linux")
}
