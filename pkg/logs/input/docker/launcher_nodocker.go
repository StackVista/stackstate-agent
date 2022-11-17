// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !docker
// +build !docker

package docker

import (
	"time"

	"github.com/StackVista/stackstate-agent/pkg/logs/auditor"
	"github.com/StackVista/stackstate-agent/pkg/logs/config"
	"github.com/StackVista/stackstate-agent/pkg/logs/pipeline"
	"github.com/StackVista/stackstate-agent/pkg/logs/service"
	"github.com/StackVista/stackstate-agent/pkg/util/retry"
)

// Launcher is not supported on non docker environment
type Launcher struct{}

// NewLauncher returns a new Launcher
func NewLauncher(readTimeout time.Duration, psources *config.LogSources, services *service.Services, pipelineProvider pipeline.Provider, registry auditor.Registry, tailFromFile, forceTailingFromFile bool) *Launcher {
	return &Launcher{}
}

// IsAvailable retrurns false - not available
func IsAvailable() (bool, *retry.Retrier) {
	return false, nil
}

// Start does nothing
func (l *Launcher) Start() {}

// Stop does nothing
func (l *Launcher) Stop() {}
