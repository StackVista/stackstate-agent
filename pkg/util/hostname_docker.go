// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build linux || windows || darwin
// +build linux windows darwin

// I don't think windows and darwin can actually be docker hosts
// but keeping it this way for build consistency (for now)

package util

import (
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/hostname"
	"github.com/StackVista/stackstate-agent/pkg/util/hostname/validate"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

func getContainerHostname() (bool, string) {

	// Cluster-agent logic: Kube apiserver
	if getKubeHostname, found := hostname.ProviderCatalog["kube_apiserver"]; found {
		log.Debug("GetHostname trying Kubernetes trough API server...")
		data, err := getKubeHostname()
		if err == nil && validate.ValidHostname(data.Hostname) == nil {
			return true, data.Hostname
		}
	}

	if config.IsContainerized() == false {
		return false, ""
	}

	// Node-agent logic: docker or kubelet

	// Docker
	log.Debug("GetHostname trying Docker API...")
	if getDockerHostnameData, found := hostname.ProviderCatalog["docker"]; found {
		data, err := getDockerHostnameData()
		if err == nil && validate.ValidHostname(data.Hostname) == nil {
			return true, data.Hostname
		}
	}

	if config.IsKubernetes() == false {
		return false, ""
	}
	// Kubelet
	if getKubeletHostnameData, found := hostname.ProviderCatalog["kubelet"]; found {
		log.Debug("GetHostname trying Kubernetes trough kubelet API...")
		data, err := getKubeletHostnameData()
		if err == nil && validate.ValidHostname(data.Hostname) == nil {
			return true, data.Hostname
		}
	}
	return false, ""
}
