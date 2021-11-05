// Package topology is responsible for gathering topology for containers
// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/util/containers"
)

const (
	containerType = "container"
)

// ContainerTopology is an interface for types capable of building container topology
type ContainerTopology interface {
	BuildContainerTopology(du *containers.ContainerUtil) error
}
