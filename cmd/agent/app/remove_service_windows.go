// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package app

import (
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func init() {
	AgentCmd.AddCommand(removesvcCommand)
}

var removesvcCommand = &cobra.Command{
	Use:   "remove-service",
	Short: "Removes the agent from the service control manager",
	Long:  ``,
	RunE:  removeService,
}

func removeService(cmd *cobra.Command, args []string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(config.ServiceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", config.ServiceName)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(config.ServiceName)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}
