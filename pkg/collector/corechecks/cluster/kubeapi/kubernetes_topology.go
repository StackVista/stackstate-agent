// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
// +build kubeapiserver

package kubeapi

import (
	"k8s.io/api/core/v1"

	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	core "github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"gopkg.in/yaml.v2"
)

const (
	kubernetesAPITopologyCheckName = "kubernetes_api_topology"
)

// TopologyConfig is the config of the API server.
type TopologyConfig struct {
	CollectTopology             bool     `yaml:"collect_topology"`
}

// TopologyCheck grabs events from the API server.
type TopologyCheck struct {
	CommonCheck
	instance              *TopologyConfig
}

func (c *TopologyConfig) parse(data []byte) error {
	// default values
	c.CollectTopology = config.Datadog.GetBool("collect_kubernetes_topology")

	return yaml.Unmarshal(data, c)
}

// Configure parses the check configuration and init the check.
func (k *TopologyCheck) Configure(config, initConfig integration.Data) error {
	err := k.ConfigureKubeApiCheck(config)

	err = k.instance.parse(config)
	if err != nil {
		_ = log.Error("could not parse the config for the API topology check")
		return err
	}

	log.Debugf("Running config %s", config)
	return nil
}

// Run executes the check.
func (k *TopologyCheck) Run() error {

	// Running the event collection.
	if !k.instance.CollectTopology {
		return nil
	}

	// initialize kube api check
	err := k.InitKubeApiCheck()
	if err != nil {
		return err
	}

	// start the topology snapshot with the batch-er
	var instance = topology.Instance{Type: "kubernetes", URL: k.KubeAPIServerHostname}
	batcher.GetBatcher().SubmitStartSnapshot(kubernetesAPITopologyCheckName, instance)

	// get all the nodes
	_, _ = k.getAllNodes()

	// get all the pods
	_, _ = k.getAllPods()

	// get all the services
	_, _ = k.getAllServices()

	// get all the containers
	batcher.GetBatcher().SubmitStopSnapshot(kubernetesAPITopologyCheckName, instance)
	batcher.GetBatcher().SubmitComplete(kubernetesAPITopologyCheckName)

	return nil
}

// get all the nodes in the k8s cluster
func (k *TopologyCheck) getAllNodes() ([]v1.Node, error) {
	nodes, err := k.ac.GetNodes()

	// map nodes
	log.Debugf("Found the following nodes: %v", nodes)

	return nodes, err
}

// get all the pods in the k8s cluster
func (k *TopologyCheck) getAllPods() ([]v1.Pod, error) {
	pods, err := k.ac.GetPods()

	// map pods
	log.Debugf("Found the following nodes: %v", pods)

	return pods, err
}

// get all the services in the k8s cluster
func (k *TopologyCheck) getAllServices() ([]v1.Service, error) {
	services, err := k.ac.GetServices()

	// map services
	log.Debugf("Found the following nodes: %v", services)

	return services, err
}

// KubernetesASFactory is exported for integration testing.
func KubernetesApiTopologyFactory() check.Check {
	return &TopologyCheck{
		CommonCheck: CommonCheck{
			CheckBase: core.NewCheckBase(kubernetesAPITopologyCheckName),
		},
		instance:  &TopologyConfig{},
	}
}


func init() {
	core.RegisterCheck(kubernetesAPITopologyCheckName, KubernetesApiEventsFactory)
}
