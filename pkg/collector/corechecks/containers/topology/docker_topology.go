// Package topology is responsible for gathering topology for containers
// StackState

//go:build docker
// +build docker

package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	dockerTopologyCheckName = "docker_topology"
)

// DockerTopologyCollector contains the checkID and topology instance for the docker topology check
//type DockerTopologyCollector struct {
//	corechecks.CheckTopologyCollector
//}

// MakeDockerTopologyCollector returns a new instance of DockerTopologyCollector
func MakeDockerTopologyCollector() *ContainerTopologyCollector {
	return &ContainerTopologyCollector{
		corechecks.MakeCheckTopologyCollector(dockerTopologyCheckName, topology.Instance{
			Type: "docker",
			URL:  "agents",
		}),
	}
}
