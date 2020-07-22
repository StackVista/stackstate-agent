// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package app

import (
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/spf13/cobra"
	"github.com/StackVista/stackstate-agent/cmd/agent/api"
	"github.com/StackVista/stackstate-agent/cmd/agent/common"
	"github.com/StackVista/stackstate-agent/cmd/agent/gui"
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/api/healthprobe"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/embed/jmx"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/dogstatsd"
	"github.com/StackVista/stackstate-agent/pkg/forwarder"
	"github.com/StackVista/stackstate-agent/pkg/logs"
	"github.com/StackVista/stackstate-agent/pkg/metadata"
	"github.com/StackVista/stackstate-agent/pkg/metadata/host"
	"github.com/StackVista/stackstate-agent/pkg/pidfile"
	"github.com/StackVista/stackstate-agent/pkg/serializer"
	"github.com/StackVista/stackstate-agent/pkg/status/health"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/StackVista/stackstate-agent/pkg/version"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/embed"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/net"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/system"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/metadata"
	_ "github.com/StackVista/stackstate-agent/pkg/metadata"
)

var (
	startCmd = &cobra.Command{
		Use:        "start",
		Deprecated: "Use \"run\" instead to start the Agent",
		RunE:       start,
	}
)

func init() {
	// attach the command to the root
	AgentCmd.AddCommand(startCmd)

	// local flags
	startCmd.Flags().StringVarP(&pidfilePath, "pidfile", "p", "", "path to the pidfile")
}

// Start the main loop
func start(cmd *cobra.Command, args []string) error {
	return run(cmd, args)
}
