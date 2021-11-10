// StackState

//go:build containerd
// +build containerd

package topology

import (
	"errors"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/containerd"
)

const (
	containerdTopologyCheckName = "containerd_topology"
)

// ContainerdTopologyCollector contains the checkID and topology instance for the docker topology check
type ContainerdTopologyCollector struct {
	corechecks.CheckTopologyCollector
	containerTopology
}

// MakeContainerdTopologyCollector returns a new instance of ContainerdTopologyCollector
func MakeContainerdTopologyCollector() *ContainerdTopologyCollector {
	return &ContainerdTopologyCollector{
		corechecks.MakeCheckTopologyCollector(containerdTopologyCheckName, topology.Instance{
			Type: "containerd",
			URL:  "agents",
		}),
		containerTopology{},
	}
}

// BuildContainerTopology collects all docker container topology
func (contd *ContainerdTopologyCollector) BuildContainerTopology(cu *containerd.ContainerdUtil) error {
	sender := batcher.GetBatcher()
	if sender == nil {
		return errors.New("no batcher instance available, skipping BuildContainerTopology")
	}

	// collect all containers as topology components
	containerComponents, err := contd.collectContainers(cu)
	if err != nil {
		return err
	}

	// submit all collected topology components
	for _, component := range containerComponents {
		sender.SubmitComponent(contd.CheckID, contd.TopologyInstance, *component)
	}

	sender.SubmitComplete(contd.CheckID)

	return nil
}

// collectContainers collects containers from the docker util and produces topology.Component
func (contd *ContainerdTopologyCollector) collectContainers(cu *containerd.ContainerdUtil) ([]*topology.Component, error) {
	cList, err := cu.GetContainers()
	if err != nil {
		return nil, err
	}

	containerComponents := contd.MapContainersToComponents(cList)

	return containerComponents, nil
}
