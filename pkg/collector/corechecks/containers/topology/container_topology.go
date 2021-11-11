// Package topology is responsible for gathering topology for containers
// StackState
package topology

import (
	"errors"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/spec"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	containerType = "container"
)

// ContainerTopologyCollector contains the checkID and topology instance for the container topology checks
type ContainerTopologyCollector struct {
	corechecks.CheckTopologyCollector
}

// BuildContainerTopology collects all docker container topology
func (ctc *ContainerTopologyCollector) BuildContainerTopology(containerUtil spec.ContainerUtil) error {
	sender := batcher.GetBatcher()
	if sender == nil {
		return errors.New("no batcher instance available, skipping BuildContainerTopology")
	}

	// collect all containers as topology components
	containerComponents, err := ctc.collectContainers(containerUtil)
	if err != nil {
		return err
	}

	// submit all collected topology components
	for _, component := range containerComponents {
		sender.SubmitComponent(ctc.CheckID, ctc.TopologyInstance, *component)
	}

	sender.SubmitComplete(ctc.CheckID)

	return nil
}

// MapContainerDataToTopologyData takes a spec.Container as input and outputs topology.Data
func (ctc *ContainerTopologyCollector) MapContainerDataToTopologyData(container *spec.Container) topology.Data {
	return topology.Data{
		"type":        container.Runtime,
		"containerID": container.ID,
		"name":        container.Name,
		"image":       container.Image,
		"mounts":      container.Mounts,
		"state":       container.State,
	}
}

// MapContainerToComponent Maps a single spec.Container to a single topology.Component
func (ctc *ContainerTopologyCollector) MapContainerToComponent(container *spec.Container) *topology.Component {
	output := &topology.Component{
		ExternalID: fmt.Sprintf("urn:%s:/%s", containerType, container.ID),
		Type: topology.Type{
			Name: containerType,
		},
		Data: ctc.MapContainerDataToTopologyData(container),
	}
	return output
}

// MapContainersToComponents Maps a slice of spec.Container(s) to a slice of topology.Component(s)
func (ctc *ContainerTopologyCollector) MapContainersToComponents(containers []*spec.Container) []*topology.Component {
	output := make([]*topology.Component, 0, len(containers))

	for _, container := range containers {
		top := ctc.MapContainerToComponent(container)
		output = append(output, top)
	}

	return output
}

// collectContainers collects containers and produces topology.Component
func (ctc *ContainerTopologyCollector) collectContainers(containerUtil spec.ContainerUtil) ([]*topology.Component, error) {
	cList, err := containerUtil.GetContainers()
	if err != nil {
		return nil, err
	}

	containerComponents := ctc.MapContainersToComponents(cList)

	return containerComponents, nil
}
