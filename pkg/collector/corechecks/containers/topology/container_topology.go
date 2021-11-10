// Package topology is responsible for gathering topology for containers
// StackState
package topology

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/spec"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	containerType = "container"
)

type containerTopology struct {
	Topology ContainerTopology
}

// ContainerTopology is an interface for types capable of building container topology
type ContainerTopology interface {
	BuildContainerTopology(du *spec.ContainerUtil) error
}

// MapContainerDataToTopologyData takes a spec.Container as input and outputs topology.Data
func (ct *containerTopology) MapContainerDataToTopologyData(container *spec.Container) topology.Data {
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
func (ct *containerTopology) MapContainerToComponent(container *spec.Container) *topology.Component {
	output := &topology.Component{
		ExternalID: fmt.Sprintf("urn:%s:/%s", containerType, container.ID),
		Type: topology.Type{
			Name:"container",
		},
		Data: ct.MapContainerDataToTopologyData(container),
	}
	return output
}
// MapContainersToComponents Maps a slice of spec.Container(s) to a slice of topology.Component(s)
func (ct *containerTopology) MapContainersToComponents(containers []*spec.Container) []*topology.Component {
	output := make([]*topology.Component, 0, len(containers))

	for _, container := range containers {
		top := ct.MapContainerToComponent(container)
		output = append(output, top)
	}

	return output
}
