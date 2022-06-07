// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build !windows && kubeapiserver
// +build !windows,kubeapiserver

package app

import (
	"github.com/StackVista/stackstate-agent/cmd/security-agent/common"
	"github.com/spf13/cobra"
)

var (
	complianceCmd = &cobra.Command{
		Use:   "compliance",
		Short: "Compliance utility commands",
	}
)

func init() {
	complianceCmd.AddCommand(common.CheckCmd(&confPath))
	ClusterAgentCmd.AddCommand(complianceCmd)
}
