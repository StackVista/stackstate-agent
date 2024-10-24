// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver && clusterchecks
// +build kubeapiserver,clusterchecks

package app

import "github.com/StackVista/stackstate-agent/cmd/cluster-agent/commands"

func init() {
	clusterChecksCmd := commands.GetClusterChecksCobraCmd(&flagNoColor, &confPath, loggerName)
	clusterChecksCmd.AddCommand(commands.RebalanceClusterChecksCobraCmd(&flagNoColor, &confPath, loggerName))

	ClusterAgentCmd.AddCommand(clusterChecksCmd)
}
