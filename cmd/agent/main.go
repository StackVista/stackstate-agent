// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows && !android
// +build !windows,!android

package main

import (
	"os"

	"github.com/StackVista/stackstate-agent/cmd/agent/app"
	// sts - import dockerswarm to load check
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/dockerswarm"
)

func main() {
	// Invoke the Agent
	if err := app.AgentCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
