// Package topology is responsible for gathering topology for containers
// StackState

//go:build cri
// +build cri

package topology

import (
	"errors"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/containers/cri"
)

const (
	criTopologyCheckName = "cri_topology"
)

// CRITopologyCollector contains the checkID and topology instance for the docker topology check
type CRITopologyCollector struct {
	corechecks.CheckTopologyCollector
}

// MakeCRITopologyCollector returns a new instance of CRITopologyCollector
func MakeCRITopologyCollector() *CRITopologyCollector {
	return &CRITopologyCollector{
		corechecks.MakeCheckTopologyCollector(criTopologyCheckName, topology.Instance{
			Type: "cri",
			URL:  "agents",
		}),
	}
}

// BuildContainerTopology collects all docker container topology
func (critc *CRITopologyCollector) BuildContainerTopology(cu *cri.CRIUtil) error {
	sender := batcher.GetBatcher()
	if sender == nil {
		return errors.New("no batcher instance available, skipping BuildContainerTopology")
	}

	// collect all containers as topology components
	containerComponents, err := critc.collectContainers(cu)
	if err != nil {
		return err
	}

	// submit all collected topology components
	for _, component := range containerComponents {
		sender.SubmitComponent(critc.CheckID, critc.TopologyInstance, *component)
	}

	sender.SubmitComplete(critc.CheckID)

	return nil
}

// collectContainers collects containers from the docker util and produces topology.Component
func (critc *CRITopologyCollector) collectContainers(cu *cri.CRIUtil) ([]*topology.Component, error) {
	cList, err := cu.GetContainers()
	if err != nil {
		return nil, err
	}

	containerComponents := make([]*topology.Component, 0)
	for _, ctr := range cList {
		containerComponent := &topology.Component{
			ExternalID: fmt.Sprintf("urn:%s:/%s", containerType, ctr.ID),
			Type:       topology.Type{Name: containerType},
			Data: topology.Data{
				"type":        ctr.Runtime,
				"containerID": ctr.ID,
				"name":        ctr.Name,
				"image":       ctr.Image,
				"mounts":      ctr.Mounts,
				"state":       ctr.State,
			},
		}

		containerComponents = append(containerComponents, containerComponent)
	}

	return containerComponents, nil
}
