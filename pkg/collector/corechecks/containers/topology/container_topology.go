// Package topology is responsible for gathering topology for containers
// StackState
package topology

import (
	"context"
	"errors"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/spec"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

const (
	containerType = "container"
)

// ContainerTopologyCollector contains the checkID and topology instance for the container topology checks
type ContainerTopologyCollector struct {
	corechecks.CheckTopologyCollector
	Hostname string
	Runtime  string
}

// MakeContainerTopologyCollector returns a new instance of DockerTopologyCollector
func MakeContainerTopologyCollector(runtime string) *ContainerTopologyCollector {
	hostname, err := util.GetHostname(context.TODO())
	if err != nil {
		log.Warnf("Can't get hostname from container collector, containers ExternalIDs will not have it: %s", err)
	}
	return &ContainerTopologyCollector{
		CheckTopologyCollector: corechecks.MakeCheckTopologyCollector(
			check.ID(fmt.Sprintf("%s_topology", containerType)), topology.Instance{
				Type: containerType,
				URL:  "agents",
			}),
		Hostname: hostname,
		Runtime:  runtime,
	}
}

// BuildContainerTopology collects all docker container topology
func (ctc *ContainerTopologyCollector) BuildContainerTopology(containerUtil spec.ContainerUtil) error {
	log.Infof("Running container topology collector for '%s' runtime", ctc.Runtime)
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
	data := topology.Data{
		"type":        container.Runtime,
		"containerId": container.ID,
		"name":        container.Name,
		"image":       container.Image,
		"mounts":      container.Mounts,
		"state":       container.State,
		"labels":      []string{runtimeLabel(container.Runtime)},
	}
	processAgentIdentifier, err := ctc.buildProcessAgentContainerIdentifier(container.ID)
	if err != nil {
		log.Warnf("Could not build process agent identifier for container: %s", err.Error())
	} else {
		data["identifiers"] = []string{processAgentIdentifier}
	}
	return data
}

// MapContainerToComponent Maps a single spec.Container to a single topology.Component
func (ctc *ContainerTopologyCollector) MapContainerToComponent(container *spec.Container) *topology.Component {
	output := &topology.Component{
		ExternalID: ctc.buildContainerExternalID(container),
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
	cList, err := containerUtil.GetContainers(context.TODO())
	if err != nil {
		return nil, err
	}

	containerComponents := ctc.MapContainersToComponents(cList)

	return containerComponents, nil
}

func (ctc *ContainerTopologyCollector) buildContainerExternalID(container *spec.Container) string {
	if ctc.Hostname == "" {
		return fmt.Sprintf("urn:%s:%s:/%s", containerType, container.Runtime, container.ID)
	}
	return fmt.Sprintf("urn:%s:%s:/%s:%s", containerType, container.Runtime, ctc.Hostname, container.ID)
}

// buildProcessAgentContainerIdentifier creates an identifier with the same format as in the process-agent
// It is added to make sure the container component from the node agent merges with the one from the process agent.
func (ctc *ContainerTopologyCollector) buildProcessAgentContainerIdentifier(containerID string) (string, error) {
	if ctc.Hostname == "" {
		return "", fmt.Errorf("no hostname found, it's not possible to build the process-agent identifier")
	}
	return fmt.Sprintf("urn:%s:/%s:%s", containerType, ctc.Hostname, containerID), nil
}

func runtimeLabel(runtime string) string {
	return fmt.Sprintf("runtime:%s", runtime)
}
