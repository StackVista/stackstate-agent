// Package topology is responsible for gathering topology for containers
// StackState

//go:build cri
// +build cri

package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	criTopologyCheckName = "cri_topology"
)

// CRITopologyCollector contains the checkID and topology instance for the docker topology check
//type CRITopologyCollector struct {
//	corechecks.CheckTopologyCollector
//}

// MakeCRITopologyCollector returns a new instance of CRITopologyCollector
func MakeCRITopologyCollector() *ContainerTopologyCollector {
	return &ContainerTopologyCollector{
		corechecks.MakeCheckTopologyCollector(criTopologyCheckName, topology.Instance{
			Type: "cri",
			URL:  "agents",
		}),
	}
}
