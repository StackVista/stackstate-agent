// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver

package kubeapi

import (
	"errors"
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// CommonCheck.
type CommonCheck struct {
	core.CheckBase
	KubeAPIServerHostname string
	ac                    *apiserver.APIClient
}

func (k *CommonCheck) ConfigureKubeAPICheck(senderManager sender.SenderManager, checkManager handler.CheckManager, integrationConfigDigest uint64, config, initConfig integration.Data, source string) error {
	return k.CommonConfigure(senderManager, checkManager, integrationConfigDigest, initConfig, config, source)
}

func (k *CommonCheck) InitKubeAPICheck() error {
	if config.Datadog.GetBool("cluster_agent.enabled") {
		var errMsg = "cluster agent is enabled. Not running Kubernetes API Server check or collecting Kubernetes Events"
		log.Debug(errMsg)
		return errors.New(errMsg)
	}

	var err error
	// API Server client initialisation on first run
	if k.ac == nil {
		// We start the API Server Client.
		k.ac, err = apiserver.GetAPIClient()
		if err != nil {
			_ = k.Warnf("Could not connect to cluster API Server: %s", err.Error())
			return err
		}
	}

	return nil
}
