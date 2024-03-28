// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows
// +build !windows

package main

import (
	"os"

	_ "expvar"         // Blank import used because this isn't directly used in this file
	_ "net/http/pprof" // Blank import used because this isn't directly used in this file

	"github.com/StackVista/stackstate-agent/pkg/util/flavor"

	"github.com/StackVista/stackstate-agent/cmd/security-agent/app"
)

func main() {
	// set the Agent flavor
	flavor.SetFlavor(flavor.SecurityAgent)

	if err := app.SecurityAgentCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
