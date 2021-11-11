// StackState

//go:build containerd
// +build containerd

package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	containerdTopologyCheckName = "containerd_topology"
)

//// ContainerdTopologyCollector contains the checkID and topology instance for the docker topology check
//type ContainerdTopologyCollector struct {
//	corechecks.CheckTopologyCollector
//}

// MakeContainerdTopologyCollector returns a new instance of ContainerdTopologyCollector
func MakeContainerdTopologyCollector() *ContainerTopologyCollector {
	return &ContainerTopologyCollector{
		corechecks.MakeCheckTopologyCollector(containerdTopologyCheckName, topology.Instance{
			Type: "containerd",
			URL:  "agents",
		}),
	}
}
