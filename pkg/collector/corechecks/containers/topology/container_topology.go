// Package topology is responsible for gathering topology for containers
// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/container_runtime"
)

const (
	containerType = "container"
)

// ContainerTopology is an interface for types capable of building container topology
type ContainerTopology interface {
	BuildContainerTopology(du *container_runtime.ContainerUtil) error
}
