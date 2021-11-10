// Package topology is responsible for gathering topology for containers
// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/spec"
)

const (
	containerType = "container"
)

// ContainerTopology is an interface for types capable of building container topology
type ContainerTopology interface {
	BuildContainerTopology(du *spec.ContainerUtil) error
}
