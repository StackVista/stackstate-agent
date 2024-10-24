// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver && !kubelet
// +build kubeapiserver,!kubelet

package apiserver

import (
	"context"

	a "github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/clustername"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

func HostnameProvider(ctx context.Context, options map[string]interface{}) (string, error) {
	nodeName, err := a.HostNodeName(ctx)
	if err != nil {
		return "", err
	}

	clusterName := clustername.GetClusterName(ctx, nodeName)
	if clusterName == "" {
		log.Debugf("Now using plain kubernetes nodename as an alias: no cluster name was set and none could be autodiscovered")
		return nodeName, nil
	}

	return nodeName + "-" + clusterName, nil
}
