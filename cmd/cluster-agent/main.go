// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !windows && kubeapiserver
// +build !windows,kubeapiserver

//go:generate go run ../../pkg/config/render_config.go dca ../../pkg/config/config_template.yaml ../../Dockerfiles/cluster-agent/datadog-cluster.yaml

package main

import (
	"os"

	_ "expvar"         // Blank import used because this isn't directly used in this file
	_ "net/http/pprof" // Blank import used because this isn't directly used in this file

	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/ksm"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/kubeapi" // sts
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/kubernetesapiserver"
	// [STS] avoid running the orchestrator. Re-enable once upstream merging has been done (if needed)
	//_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/orchestrator"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/net"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/system/cpu"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/system/disk"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/system/filehandles"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/system/memory"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/system/uptime"
	_ "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/system/winproc"
	"github.com/StackVista/stackstate-agent/pkg/util/flavor"
	"github.com/StackVista/stackstate-agent/pkg/util/log"

	"github.com/StackVista/stackstate-agent/cmd/cluster-agent/app"
)

func main() {
	// set the Agent flavor
	flavor.SetFlavor(flavor.ClusterAgent)
	if err := app.ClusterAgentCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}
