// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017-present Datadog, Inc.

// +build docker

package collectors

import (
	"context"
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/util/containers"
	ecsutil "github.com/StackVista/stackstate-agent/pkg/util/ecs"
)

const (
	ecsFargateCollectorName = "ecs_fargate"
)

// ECSFargateCollector gets container list and metrics from the ecs metadata api.
type ECSFargateCollector struct{}

// Detect tries to connect to the ECS metadata API
func (c *ECSFargateCollector) Detect() error {
	if ecsutil.IsFargateInstance(context.TODO()) {
		return nil
	}
	return fmt.Errorf("failed to connect to task metadata API")
}

// List gets all running containers
func (c *ECSFargateCollector) List() ([]*containers.Container, error) {
	return ecsutil.ListContainersInCurrentTask()
}

// UpdateMetrics updates metrics on an existing list of containers
func (c *ECSFargateCollector) UpdateMetrics(cList []*containers.Container) error {
	return ecsutil.UpdateContainerMetrics(cList)
}

func ecsFargateFactory() Collector {
	return &ECSFargateCollector{}
}

func init() {
	registerCollector(ecsFargateCollectorName, ecsFargateFactory, NodeOrchestrator)
}
