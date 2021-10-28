// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build kubeapiserver && !kubelet
// +build kubeapiserver,!kubelet

package apiserver

import (
	"github.com/StackVista/stackstate-agent/pkg/util/hostname/hostnamedata"
	a "github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/clustername"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

func HostnameProvider() (*hostnamedata.HostnameData, error) {
	nodeName, err := a.HostNodeName()
	if err != nil {
		return nil, err
	}

	clusterName := clustername.GetClusterName()
	if clusterName == "" {
		log.Debugf("Now using plain kubernetes nodename as an alias: no cluster name was set and none could be autodiscovered")
		return hostnamedata.JustHostname(nodeName, nil)
	}

	return hostnamedata.JustHostname(nodeName+"-"+clusterName, nil)
}
