// Package topology is responsible for gathering topology for containers
// StackState

//go:build docker
// +build docker

package topology

import (
	"errors"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/docker"
)

const (
	dockerTopologyCheckName = "docker_topology"
)

// DockerTopologyCollector contains the checkID and topology instance for the docker topology check
type DockerTopologyCollector struct {
	corechecks.CheckTopologyCollector
}

// MakeDockerTopologyCollector returns a new instance of DockerTopologyCollector
func MakeDockerTopologyCollector() *DockerTopologyCollector {
	return &DockerTopologyCollector{
		corechecks.MakeCheckTopologyCollector(dockerTopologyCheckName, topology.Instance{
			Type: "docker",
			URL:  "agents",
		}),
	}
}

// BuildContainerTopology collects all docker container topology
func (dt *DockerTopologyCollector) BuildContainerTopology(du *docker.DockerUtil) error {
	sender := batcher.GetBatcher()
	if sender == nil {
		return errors.New("no batcher instance available, skipping BuildContainerTopology")
	}

	// collect all containers as topology components
	containerComponents, err := dt.collectContainers(du)
	if err != nil {
		return err
	}

	// submit all collected topology components
	for _, component := range containerComponents {
		sender.SubmitComponent(dt.CheckID, dt.TopologyInstance, *component)
	}

	sender.SubmitComplete(dt.CheckID)

	return nil
}

// collectContainers collects containers from the docker util and produces topology.Component
func (dt *DockerTopologyCollector) collectContainers(du *docker.DockerUtil) ([]*topology.Component, error) {
	cList, err := du.GetContainers()
	if err != nil {
		return nil, err
	}

	ct := containerTopology{}

	containerComponents := ct.MapContainersToComponents(cList)

	return containerComponents, nil
}
