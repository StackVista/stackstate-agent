// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

//go:build docker
// +build docker

package hostname

import (
	"context"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/clustername"
)

func init() {
	if config.IsDockerSwarm() == true {
		RegisterHostnameProvider("dockerswarm", func(ctx context.Context, options map[string]interface{}) (string, error) {
			return clustername.GetClusterName(ctx, ""), nil
		})
	}
}
